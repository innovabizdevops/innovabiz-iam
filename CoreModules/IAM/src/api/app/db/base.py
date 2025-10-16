"""
INNOVABIZ IAM - Base de Dados
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Definição da classe base para modelos SQLAlchemy
"""

from typing import Any
from sqlalchemy.ext.declarative import as_declarative, declared_attr
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import sessionmaker


@as_declarative()
class Base:
    """
    Classe base para todos os modelos SQLAlchemy.
    
    Implementa funcionalidades comuns como:
    - Geração automática de nome de tabela
    - Métodos comuns para serialização
    - Métodos utilitários para operações de banco
    """
    
    id: Any
    __name__: str
    
    # Gera nome de tabela automaticamente baseado no nome da classe
    @declared_attr
    def __tablename__(cls) -> str:
        # Convert CamelCase para snake_case para nomes de tabela
        # Ex: UserProfile -> user_profile
        result = ""
        for index, char in enumerate(cls.__name__):
            if char.isupper() and index > 0:
                result += "_"
            result += char.lower()
        return result
    
    def dict(self) -> dict:
        """
        Converte a entidade em um dicionário.
        Útil para serialização e APIs.
        
        Returns:
            dict: A entidade em formato de dicionário
        """
        return {
            column.name: getattr(self, column.name)
            for column in self.__table__.columns
            if hasattr(self, column.name)
        }