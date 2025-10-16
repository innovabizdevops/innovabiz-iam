import uvicorn
import os
from dotenv import load_dotenv
from app.core.logger import logger
from app.core.config import settings
from app.db.session import check_db_connection

# Carrega variáveis de ambiente do arquivo .env
load_dotenv()

if __name__ == "__main__":
    try:
        # Verifica a conexão com o banco de dados antes de iniciar o servidor
        if not check_db_connection():
            logger.error("Não foi possível conectar ao banco de dados. Verifique as configurações.")
            exit(1)
        
        logger.info(f"Iniciando servidor IAM API v{settings.API_VERSION}")
        logger.info(f"Ambiente: {'Produção' if settings.PRODUCTION else 'Desenvolvimento'}")
        logger.info(f"Servidor: {settings.HOST}:{settings.PORT}")
        
        # Inicia o servidor uvicorn
        uvicorn.run(
            "app.main:app",
            host=settings.HOST,
            port=settings.PORT,
            reload=settings.DEBUG,
            log_level=settings.LOG_LEVEL.lower()
        )
    except Exception as e:
        logger.critical(f"Erro ao iniciar o servidor: {str(e)}")
        exit(1)
