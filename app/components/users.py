from flask import Blueprint, jsonify, request, current_app
from models.models import db_session, Users
from flask_jwt_extended import create_access_token, jwt_required, get_jwt_identity
from hashlib import sha256
from flasgger import swag_from
from uuid_extensions import uuid7str
from datetime import datetime
import secrets


users = Blueprint("users", __name__)
URL_PREFIX = "/api"

@users.route(f"{URL_PREFIX}/users/register")
@swag_from("../swagger_yaml/users_register.yml")
def user_register():
    data = request.json
    username = data["username"]
    password = data["password"]

    user_query = Users.query.filter_by(username = username).first()
    if user_query:
        return jsonify({"status": "error", "reason": "username already exists"}), 400
    
    id = uuid7str()
    salt = secrets.token_hex(32)
    hashed_password = sha256((username + password + salt).encode("UTF-8")).hexdigest()

    try:
        user = Users(id=id, username=username, hashed_password=hashed_password, salt=salt)
        db_session.add(user)
        db_session.commit()
        access_token = create_access_token(identity=id)
    except Exception as e:
        db_session.rollback()
        return jsonify({"status": "error", "reason": str(e)}), 500
    finally:
        db_session.close()

    return jsonify({"status": "success", "access_token": access_token, "username": username}), 201
