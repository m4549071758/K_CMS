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
    hashed_password = sha256((id + password + salt).encode("UTF-8")).hexdigest()

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

@users.route(f"{URL_PREFIX}/users/update", methods=["PATCH"])
@jwt_required
@swag_from("../swagger_yaml/users_update.yml")
def user_update():
    current_user = get_jwt_identity()
    data = request.json

    target = data["target"]
    
    if target == "username":
        new_username = data["new_username"]
        user_query = Users.query.filter_by(username=new_username).first()
        if user_query:
            return jsonify({"status": "error", "reason": "username already exists"}), 400
        current_user_query = Users.query.filter_by(id=current_user).first()
        current_user_query.username = new_username
    elif target == "password":
        current_password = data["current_password"]
        new_password = data["new_password"]

        current_user_query = Users.query.filter_by(id=current_user).first()

        current_hashed_password = sha256((current_user_query.id + current_password + current_user_query.salt).encode("UTF-8")).hexdigest()
        if current_hashed_password != current_user_query.hashed_password:
            return jsonify({"status": "error", "reason": "current password is incorrect"}), 400
        
        salt = secrets.token_hex(32)
        hashed_password = sha256((current_user_query.id + new_password + salt).encode("UTF-8")).hexdigest()
        current_user_query.hashed_password = hashed_password
        current_user_query.salt = salt
    else:
        return jsonify({"status": "error", "reason": "target is invalid"}), 400
    
    try:
        db_session.commit()
    except Exception as e:
        db_session.rollback()
        return jsonify({"status": "error", "reason": str(e)}), 500
    finally:
        db_session.close()

@users.route(f"{URL_PREFIX}/users/delete", methods=["DELETE"])
@jwt_required
@swag_from("../swagger_yaml/users_delete.yml")
def user_delete():
    current_user = get_jwt_identity()
    
    user_query = Users.query.filter_by(id=current_user).first()
    if not user_query:
        return jsonify({"status": "error", "reason": "user not found"}), 404
    
    try:
        db_session.delete(user_query)
        db_session.commit()
    except Exception as e:
        db_session.rollback()
        return jsonify({"status": "error", "reason": str(e)}), 500
    finally:
        db_session.close()