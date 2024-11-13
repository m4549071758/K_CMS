from flask import Flask, request, jsonify
from flask_jwt_extended import JWTManager, create_access_token, get_jwt_identity, jwt_required, create_refresh_token
from flask_json import FlaskJSON
from flasgger import Swagger
from flasgger.utils import swag_from
from hashlib import sha256
from uuid_extensions import uuid7str
from models.models import db_session
from werkzeug.middleware.proxy_fix import ProxyFix
from datetime import timedelta
import secrets

app = Flask(__name__)
app.secret_key = secrets.token_hex(24)

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

# テストAPI
@app.route(f"{URL_PREFIX}/ping", methods=["GET"])
@swag_from("swagger_yaml/ping.yml")
def ping():
    return jsonify({"message": "pong"})