"""
INNOVABIZ - Signature Service
===============================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0

Serviço de assinatura digital para certificados e documentos
de compliance, garantindo integridade e autenticidade.
===============================================================
"""

import os
import base64
import hmac
import hashlib
import logging
import datetime
import json
from typing import Dict, Any, Optional
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.serialization import load_pem_private_key
from cryptography.exceptions import InvalidSignature

# Configurar logging
logger = logging.getLogger(__name__)


class SignatureService:
    """
    Serviço de assinatura digital para garantir a integridade e
    autenticidade de documentos e certificados.
    
    Implementa dois métodos de assinatura:
    1. HMAC-SHA256 - Para assinaturas simétricas mais simples
    2. RSA-SHA256 - Para assinaturas assimétricas mais robustas
    """
    
    def __init__(
        self,
        secret_key: Optional[str] = None,
        private_key_path: Optional[str] = None,
        public_key_path: Optional[str] = None,
        signature_method: str = "HMAC"
    ):
        """
        Inicializa o serviço de assinatura.
        
        Args:
            secret_key: Chave secreta para método HMAC
            private_key_path: Caminho para a chave privada RSA
            public_key_path: Caminho para a chave pública RSA
            signature_method: Método de assinatura ('HMAC' ou 'RSA')
        """
        self.signature_method = signature_method
        self._initialize_keys(secret_key, private_key_path, public_key_path)
        logger.info(f"Serviço de assinatura inicializado: {signature_method}")
    
    def _initialize_keys(
        self,
        secret_key: Optional[str],
        private_key_path: Optional[str],
        public_key_path: Optional[str]
    ):
        """
        Inicializa as chaves de assinatura com base no método escolhido.
        
        Args:
            secret_key: Chave secreta para método HMAC
            private_key_path: Caminho para a chave privada RSA
            public_key_path: Caminho para a chave pública RSA
        """
        if self.signature_method == "HMAC":
            # Usar a chave fornecida ou gerar uma nova
            self.secret_key = secret_key or os.environ.get("SIGNATURE_SECRET_KEY")
            
            # Se ainda não tiver chave, gerar uma nova
            if not self.secret_key:
                self.secret_key = base64.b64encode(os.urandom(32)).decode('utf-8')
                logger.warning("Chave secreta HMAC gerada automaticamente. Considere fornecer uma chave persistente.")
        
        elif self.signature_method == "RSA":
            # Carregar chave privada
            self.private_key = None
            if private_key_path:
                try:
                    with open(private_key_path, "rb") as key_file:
                        key_data = key_file.read()
                        self.private_key = load_pem_private_key(
                            key_data,
                            password=None  # Considerar usar senha para proteção adicional
                        )
                except Exception as e:
                    logger.error(f"Erro ao carregar chave privada: {e}")
            
            # Carregar chave pública
            self.public_key = None
            if public_key_path:
                try:
                    with open(public_key_path, "rb") as key_file:
                        key_data = key_file.read()
                        self.public_key = serialization.load_pem_public_key(key_data)
                except Exception as e:
                    logger.error(f"Erro ao carregar chave pública: {e}")
            
            # Verificar se as chaves estão disponíveis com base na operação
            if not self.private_key:
                logger.warning("Chave privada RSA não disponível. Assinatura não será possível.")
            
            if not self.public_key:
                logger.warning("Chave pública RSA não disponível. Verificação não será possível.")
        
        else:
            raise ValueError(f"Método de assinatura não suportado: {self.signature_method}")
    
    def sign(self, content: str) -> str:
        """
        Assina o conteúdo fornecido usando o método configurado.
        
        Args:
            content: Conteúdo a ser assinado
            
        Returns:
            Assinatura em formato base64
        """
        if self.signature_method == "HMAC":
            return self._hmac_sign(content)
        elif self.signature_method == "RSA":
            return self._rsa_sign(content)
        else:
            raise ValueError(f"Método de assinatura não suportado: {self.signature_method}")
    
    def verify(self, content: str, signature: str) -> bool:
        """
        Verifica se a assinatura é válida para o conteúdo.
        
        Args:
            content: Conteúdo original
            signature: Assinatura a verificar
            
        Returns:
            True se a assinatura for válida, False caso contrário
        """
        if self.signature_method == "HMAC":
            return self._hmac_verify(content, signature)
        elif self.signature_method == "RSA":
            return self._rsa_verify(content, signature)
        else:
            raise ValueError(f"Método de assinatura não suportado: {self.signature_method}")
    
    def _hmac_sign(self, content: str) -> str:
        """
        Assina o conteúdo usando HMAC-SHA256.
        
        Args:
            content: Conteúdo a ser assinado
            
        Returns:
            Assinatura em formato base64
        """
        if not self.secret_key:
            raise ValueError("Chave secreta não configurada para assinatura HMAC")
        
        # Criar assinatura HMAC
        key = self.secret_key.encode('utf-8')
        message = content.encode('utf-8')
        
        signature = hmac.new(key, message, hashlib.sha256).digest()
        return base64.b64encode(signature).decode('utf-8')
    
    def _hmac_verify(self, content: str, signature: str) -> bool:
        """
        Verifica uma assinatura HMAC-SHA256.
        
        Args:
            content: Conteúdo original
            signature: Assinatura a verificar
            
        Returns:
            True se a assinatura for válida, False caso contrário
        """
        if not self.secret_key:
            logger.error("Chave secreta não configurada para verificação HMAC")
            return False
        
        try:
            # Decodificar assinatura
            signature_bytes = base64.b64decode(signature)
            
            # Recalcular assinatura e comparar
            key = self.secret_key.encode('utf-8')
            message = content.encode('utf-8')
            
            expected_signature = hmac.new(key, message, hashlib.sha256).digest()
            
            # Comparação de tempo constante para prevenir ataques de timing
            return hmac.compare_digest(signature_bytes, expected_signature)
        
        except Exception as e:
            logger.error(f"Erro na verificação HMAC: {e}")
            return False
    
    def _rsa_sign(self, content: str) -> str:
        """
        Assina o conteúdo usando RSA-SHA256.
        
        Args:
            content: Conteúdo a ser assinado
            
        Returns:
            Assinatura em formato base64
        """
        if not self.private_key:
            raise ValueError("Chave privada RSA não configurada para assinatura")
        
        # Assinar o conteúdo
        content_bytes = content.encode('utf-8')
        
        signature = self.private_key.sign(
            content_bytes,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )
        
        return base64.b64encode(signature).decode('utf-8')
    
    def _rsa_verify(self, content: str, signature: str) -> bool:
        """
        Verifica uma assinatura RSA-SHA256.
        
        Args:
            content: Conteúdo original
            signature: Assinatura a verificar
            
        Returns:
            True se a assinatura for válida, False caso contrário
        """
        if not self.public_key:
            logger.error("Chave pública RSA não configurada para verificação")
            return False
        
        try:
            # Decodificar assinatura
            signature_bytes = base64.b64decode(signature)
            
            # Conteúdo a verificar
            content_bytes = content.encode('utf-8')
            
            # Verificar assinatura
            self.public_key.verify(
                signature_bytes,
                content_bytes,
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH
                ),
                hashes.SHA256()
            )
            
            # Se não lançou exceção, a assinatura é válida
            return True
            
        except InvalidSignature:
            logger.warning(f"Assinatura RSA inválida para o conteúdo")
            return False
        except Exception as e:
            logger.error(f"Erro na verificação RSA: {e}")
            return False
    
    @staticmethod
    def generate_key_pair(
        output_private_key_path: str,
        output_public_key_path: str,
        key_size: int = 2048
    ) -> bool:
        """
        Gera um novo par de chaves RSA.
        
        Args:
            output_private_key_path: Caminho para salvar a chave privada
            output_public_key_path: Caminho para salvar a chave pública
            key_size: Tamanho da chave em bits
            
        Returns:
            True se gerado com sucesso, False caso contrário
        """
        try:
            # Gerar chave privada
            private_key = rsa.generate_private_key(
                public_exponent=65537,
                key_size=key_size
            )
            
            # Extrair chave pública
            public_key = private_key.public_key()
            
            # Serializar chave privada
            private_pem = private_key.private_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PrivateFormat.PKCS8,
                encryption_algorithm=serialization.NoEncryption()
            )
            
            # Serializar chave pública
            public_pem = public_key.public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            )
            
            # Salvar chaves em disco
            with open(output_private_key_path, 'wb') as f:
                f.write(private_pem)
            
            with open(output_public_key_path, 'wb') as f:
                f.write(public_pem)
            
            logger.info(f"Par de chaves RSA gerado e salvo em {output_private_key_path} e {output_public_key_path}")
            return True
            
        except Exception as e:
            logger.error(f"Erro ao gerar par de chaves RSA: {e}")
            return False


class DocumentSignatureManager:
    """
    Gerenciador de assinaturas para documentos e certificados.
    Provê funcionalidades adicionais sobre o SignatureService básico.
    """
    
    def __init__(self, signature_service: SignatureService):
        """
        Inicializa o gerenciador de assinaturas.
        
        Args:
            signature_service: Serviço de assinatura a utilizar
        """
        self.signature_service = signature_service
        self.logger = logging.getLogger(__name__ + ".DocumentSignatureManager")
    
    def sign_document(
        self,
        document: Dict[str, Any],
        issuer: str = "INNOVABIZ Authority",
        include_timestamp: bool = True
    ) -> Dict[str, Any]:
        """
        Assina um documento e inclui metadados da assinatura.
        
        Args:
            document: Documento a ser assinado
            issuer: Emissor da assinatura
            include_timestamp: Incluir timestamp na assinatura
            
        Returns:
            Documento com assinatura e metadados
        """
        # Criar cópia do documento para não modificar o original
        signed_doc = document.copy()
        
        # Remover assinatura anterior se existir
        if "signature" in signed_doc:
            del signed_doc["signature"]
            
        if "signature_metadata" in signed_doc:
            del signed_doc["signature_metadata"]
        
        # Preparar metadados da assinatura
        metadata = {
            "issuer": issuer,
            "method": self.signature_service.signature_method
        }
        
        if include_timestamp:
            metadata["timestamp"] = datetime.datetime.utcnow().isoformat()
        
        # Conteúdo a assinar - documento serializado em ordem consistente
        content = json.dumps(signed_doc, sort_keys=True)
        
        # Gerar assinatura
        signature = self.signature_service.sign(content)
        
        # Adicionar assinatura e metadados ao documento
        signed_doc["signature"] = signature
        signed_doc["signature_metadata"] = metadata
        
        return signed_doc
    
    def verify_document(self, document: Dict[str, Any]) -> bool:
        """
        Verifica a assinatura de um documento.
        
        Args:
            document: Documento a verificar
            
        Returns:
            True se a assinatura for válida, False caso contrário
        """
        # Verificar se o documento tem assinatura
        if "signature" not in document:
            self.logger.warning("Documento sem assinatura")
            return False
        
        signature = document.get("signature")
        
        # Criar cópia do documento sem a assinatura para verificação
        doc_to_verify = document.copy()
        del doc_to_verify["signature"]
        
        if "signature_metadata" in doc_to_verify:
            del doc_to_verify["signature_metadata"]
        
        # Preparar conteúdo para verificação
        content = json.dumps(doc_to_verify, sort_keys=True)
        
        # Verificar assinatura
        return self.signature_service.verify(content, signature)
    
    def add_timestamp_to_signature(self, document: Dict[str, Any]) -> Dict[str, Any]:
        """
        Adiciona timestamp atual à assinatura de um documento.
        
        Args:
            document: Documento assinado
            
        Returns:
            Documento com timestamp atualizado
        """
        if "signature" not in document or "signature_metadata" not in document:
            raise ValueError("Documento não está assinado")
        
        # Atualizar timestamp nos metadados
        document["signature_metadata"]["timestamp"] = datetime.datetime.utcnow().isoformat()
        
        # Re-assinar o documento com os novos metadados
        return self.sign_document(
            document,
            issuer=document["signature_metadata"].get("issuer", "INNOVABIZ Authority"),
            include_timestamp=False  # Já adicionamos manualmente
        )
