from sqlalchemy import create_engine, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session
from contextlib import contextmanager
from app.core.config import settings
from app.core.logger import logger

# Configura o engine do SQLAlchemy com pool de conexões
engine = create_engine(
    settings.DATABASE_URL,
    pool_size=settings.DATABASE_POOL_SIZE,
    max_overflow=settings.DATABASE_MAX_OVERFLOW,
    pool_timeout=30,  # Timeout para obter conexão do pool
    pool_recycle=1800,  # Recicla conexões a cada 30 minutos
    connect_args={"options": f"-c search_path={settings.DATABASE_SCHEMA},public"}
)

# Sessão local para uso na API
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Base para as classes de modelo
Base = declarative_base()

def get_db():
    """
    Dependency para obter uma sessão de banco de dados
    """
    db = SessionLocal()
    try:
        # Define o tenant_id ao iniciar a sessão
        yield db
    finally:
        db.close()

@contextmanager
def get_db_context():
    """
    Context manager para obter uma sessão de banco de dados
    """
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

def check_db_connection():
    """
    Verificar conexão com o banco de dados
    """
    try:
        with get_db_context() as db:
            db.execute(text("SELECT 1"))
            return True
    except Exception as e:
        logger.error(f"Database connection error: {str(e)}")
        return False
