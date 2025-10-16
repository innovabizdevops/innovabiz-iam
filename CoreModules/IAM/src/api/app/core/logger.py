import logging
import sys
from logging.handlers import RotatingFileHandler
from app.core.config import settings

# Configuração do logger
logger = logging.getLogger("innovabiz_iam_api")
logger.setLevel(getattr(logging, settings.LOG_LEVEL))

# Formatação dos logs
log_format = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)

# Handler para console
console_handler = logging.StreamHandler(sys.stdout)
console_handler.setFormatter(log_format)
logger.addHandler(console_handler)

# Handler para arquivo (opcional)
if settings.LOG_FILE:
    file_handler = RotatingFileHandler(
        settings.LOG_FILE,
        maxBytes=10485760,  # 10MB
        backupCount=10
    )
    file_handler.setFormatter(log_format)
    logger.addHandler(file_handler)

# Verificação inicial de configuração
logger.info(f"Logging initialized with level: {settings.LOG_LEVEL}")
if settings.DEBUG:
    logger.info("Application running in DEBUG mode")
