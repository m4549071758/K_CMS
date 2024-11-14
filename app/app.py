from flask import Flask, request, jsonify
from flask_jwt_extended import JWTManager, create_access_token, get_jwt_identity, jwt_required, create_refresh_token
from flask_json import FlaskJSON
from flasgger import Swagger
from flasgger.utils import swag_from
from hashlib import sha256
from uuid_extensions import uuid7str
from models.models import db_session, Posts, Users
from werkzeug.middleware.proxy_fix import ProxyFix
from datetime import timedelta
from app.components import users
import secrets
import os

app = Flask(__name__)
app.secret_key = secrets.token_hex(24)
app.register_blueprint(users.users)

URL_PREFIX = "/api"

jwt = JWTManager(app)
app.config["JWT_SECRET_KEY"] = secrets.token_hex(24)
app.config["JWT_ALGORITHM"] = "HS256"
app.config["JWT_ACCESS_TOKEN_EXPIRES"] = timedelta(minutes=30)
app.config["JWT_REFRESH_TOKEN_EXPIRES"] = timedelta(days=7)

FlaskJSON(app)
app.config["JSON_AS_ASCII"] = False

app.config["SWAGGER"] = {
    "title": "ヘッドレスCMS API",
    "uiversion": 3,
    "version": "0.0.1",
}
swagger = Swagger(app)

app.wsgi_app = ProxyFix(app.wsgi_app)

SAVE_FOLDER = "../blog/public/assets/blog"
ALLOWED_EXTENSIONS = {"png", "jpg", "jpeg", "gif", "webp"}

def jwt_unauthorized_loader_handler(reason):
    return jsonify({"status": "error", "reason": reason}), 401

@app.route(f"{URL_PREFIX}/posts/publish", methods=["POST"])
@jwt_required
@swag_from("swagger_yaml/posts_publish.yml")
def publish_post():
    current_user = get_jwt_identity()
    data = request.json

    post_id = uuid7str()
    os.makedirs(f"./blog/_posts/{post_id}", exist_ok=True)

    if "ogImage" not in data.files:
        og_image = "/assets/blog/dynamic-routing/cover.webp"
    else:
        image = data.files["ogImage"]
        os.makedirs(f"{SAVE_FOLDER}/{post_id}", exist_ok=True)
        image.save(f"{SAVE_FOLDER}/{post_id}/{image.filename}")
        # あとでwebp化処理
        og_image = f"/assets/blog/{post_id}/{image.filename}"

    title = data["title"]
    excerpt = data["excerpt"]
    tags = data["tags"]
    date = data["date"]
    markdown = data["markdown"]

    new_tags = """"""
    for tag in tags:
        tags += f"  - {tag}\n"

    new_markdown = f"""
        ---
        title: {title}
        excerpt: {excerpt}
        coverImage: {og_image}
        ogImage: 
          url: {og_image}
        tags: 
        {new_tags}
        date: {date}
        ---
        {markdown}
    """
    
    with open(f"./blog/_posts/{post_id}.md", "w") as f:
        f.write(new_markdown)
    
    posts = Posts(id=post_id, title=title, tags=tags, date=date, user_id=current_user)

    try:
        db_session.add(posts)
        db_session.commit()
        db_session.close()
        return jsonify({"message": "success"}), 200
    except Exception as e:
        db_session.rollback()
        db_session.close()
        return jsonify({"message": "Internal Server Error"}), 500


# テストAPI
@app.route(f"{URL_PREFIX}/ping", methods=["GET"])
@swag_from("swagger_yaml/ping.yml")
def ping():
    return jsonify({"message": "pong"})