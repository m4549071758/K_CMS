import signal
import sys
from app.app import app
from models.models import db_session

def signal_handler(signal, frame):
    print("exit...")
    # ここにDBなどのクローズ処理
    db_session.close()
    db_session.remove()
    sys.exit(0)

signal.signal(signal.SIGINT, signal_handler)

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8080, debug=True)