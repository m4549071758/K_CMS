from models.db_config import DB_HOST, DB_NAME, DB_USER, DB_PASSWORD
from sqlalchemy import create_engine
from sqlalchemy_utils import UUIDType
from sqlalchemy.ext.declarative import declarative_base
import os

engine = create_engine(f"mysql+mysqlconnector://{DB_USER}:{DB_PASSWORD}@{DB_HOST}/{DB_NAME}?charset=utf8mb4&collation=utf8mb4_general_ci", pool_size=50, max_overflow=20, echo=False) # echoは運用時Falseにする
db_session = scoped_session(sessionmaker(autocommit=False, autoflush=True, bind=engine))

class Base(object):
    __table_args__ = {
        "mysql_default_charset": "utf8mb4",
        "mysql_collate": "utf8mb4_general_ci"
    }

Base = declarative_base()

Base.metadata.create_all(engine)
Base.query = db_session.query_property()