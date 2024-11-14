from models.db_config import DB_HOST, DB_NAME, DB_USER, DB_PASSWORD
from sqlalchemy import create_engine, Column, CHAR, VARCHAR, TEXT, DATE, DATETIME, ForeignKey
from sqlalchemy_utils import UUIDType
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import scoped_session, sessionmaker, relationship
from datetime import datetime
import os


engine = create_engine(f"mysql+mysqlconnector://{DB_USER}:{DB_PASSWORD}@{DB_HOST}/{DB_NAME}?charset=utf8mb4&collation=utf8mb4_general_ci", pool_size=50, max_overflow=20, echo=False) # echoは運用時Falseにする
db_session = scoped_session(sessionmaker(autocommit=False, autoflush=True, bind=engine))

class Base(object):
    __table_args__ = {
        "mysql_default_charset": "utf8mb4",
        "mysql_collate": "utf8mb4_general_ci"
    }

Base = declarative_base()

class Users(Base):
    __tablename__ = "users"

    id = Column(UUIDType(binary=False), primary_key=True, index=True)
    username = Column(VARCHAR(32), unique=True, nullable=False)
    hashed_password = Column(CHAR(64), nullable=False)
    salt = Column(CHAR(64), nullable=False)
    account_type = Column(CHAR(5), nullable=False)
    createdAt = Column(DATETIME, default=datetime.now())
    updatedAt = Column(DATETIME, default=datetime.now(), onupdate=datetime.now())
    posts = relationship("Posts", backref="user", cascade="all, delete-orphan")

class Posts(Base):
    __tablename__ = "posts"

    id = Column(UUIDType(binary=False), primary_key=True)
    title = Column(VARCHAR(255), nullable=False)
    tags = Column(TEXT, nullable=False)
    date = Column(DATE, nullable=False)
    user_id = Column(ForeignKey("users.id"), nullable=False)

Base.metadata.create_all(engine)
Base.query = db_session.query_property()