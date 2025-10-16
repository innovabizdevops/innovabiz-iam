#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Consumidor de Eventos de Validação Documental

Este módulo implementa o consumidor especializado em processar eventos
de validação documental para detecção de fraudes.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import datetime
from typing import Dict, Any, List, Optional, Union
import pandas as pd
import numpy as np
from concurrent.futures import ThreadPoolExecutor

# Importação do consumidor base
try:
    from infrastructure.fraud_detection.event_consumers.base_consumer import BaseEventConsumer, ProcessingResult
except ImportError:
    import os
    import sys
    sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
    from event_consumers.base_consumer import BaseEventConsumer, ProcessingResult

# Configuração de logging
logger = logging.getLogger("document_validation_consumer")
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.addHandler(handler)
logger.setLevel(logging.INFO)


class DocumentValidationConsumer(BaseEventConsumer):
    """
    Consumidor especializado em processar eventos de validação documental.
    
    Este consumidor analisa eventos de validação de documentos em busca de padrões
    suspeitos e sinais de fraude com base em heurísticas e modelos de ML.
    """
    
    def __init__(
        self, 
        consumer_group_id: str = "document_validation_consumer", 
        config_file: Optional[str] = None,
        region_code: Optional[str] = None,
        model_path: Optional[str] = None,
        rules_path: Optional[str] = None,
        max_workers: int = 4
    ):
        """
        Inicializa o consumidor de validação documental.
        
        Args:
            consumer_group_id: ID do grupo de consumidores
            config_file: Caminho para o arquivo de configuração
            region_code: Código da região para filtrar eventos (AO, BR, MZ, PT)
            model_path: Caminho para o modelo de ML
            rules_path: Caminho para as regras de validação
            max_workers: Número máximo de workers para processamento paralelo
        """
        # Tópicos a serem consumidos
        topics = ["fraud_detection.document_validation_events"]
        
        # Se houver região específica, ajustar tópicos
        if region_code:
            region_prefixes = {
                "AO": "angola.",
                "BR": "brasil.",
                "MZ": "mocambique.",
                "PT": "portugal."
            }
            if region_code in region_prefixes:
                topics = [f"{region_prefixes[region_code]}{topic}" for topic in topics]
        
        # Inicializar consumidor base
        super().__init__(
            consumer_group_id=consumer_group_id,
            topics=topics,
            config_file=config_file,
            region_code=region_code,
            enable_auto_commit=False,  # Controle manual de commit
            isolation_level="read_committed"
        )
        
        # Variáveis específicas deste consumidor
        self.model_path = model_path
        self.rules_path = rules_path
        self.max_workers = max_workers
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        
        # Estatísticas específicas de validação documental
        self.validation_stats = {
            "total_documents": 0,
            "valid_documents": 0,
            "invalid_documents": 0,
            "suspicious_documents": 0,
            "fraud_signals": {},
            "documents_by_type": {},
            "documents_by_country": {},
            "avg_confidence_score": 0.0
        }
        
        # Carregar regras e modelos
        self.rules = self._load_rules()
        self.model = self._load_model()
        
        logger.info(f"Consumidor de validação documental inicializado para região: {region_code or 'Todas'}")
    
    def _load_rules(self):
        """
        Carrega as regras de validação documental.
        
        Returns:
            Dict: Regras de validação por tipo de documento e país
        """
        if not self.rules_path:
            # Regras padrão
            return {
                "default": {
                    "min_confidence_score": 0.7,
                    "required_steps": ["format_check", "checksum_validation"],
                    "suspicious_patterns": [
                        {"pattern": "multiple_failures", "risk": "high"},
                        {"pattern": "rapid_document_change", "risk": "high"},
                        {"pattern": "external_validation_failure", "risk": "medium"}
                    ]
                },
                "AO": {
                    "bi": {
                        "min_confidence_score": 0.75,
                        "required_steps": ["format_check", "checksum_validation", "circ_verification"],
                        "validity_period": {"min_years": 5, "max_years": 10},
                        "suspicious_patterns": [
                            {"pattern": "invalid_format", "risk": "high"},
                            {"pattern": "expired_document", "risk": "medium"},
                            {"pattern": "altered_document", "risk": "high"}
                        ]
                    }
                },
                "BR": {
                    "cpf": {
                        "min_confidence_score": 0.8,
                        "required_steps": ["format_check", "checksum_validation", "receita_verification"],
                        "suspicious_patterns": [
                            {"pattern": "sequential_digits", "risk": "medium"},
                            {"pattern": "known_invalid", "risk": "high"},
                            {"pattern": "multiple_attempts", "risk": "medium"}
                        ]
                    },
                    "cnpj": {
                        "min_confidence_score": 0.8,
                        "required_steps": ["format_check", "checksum_validation", "receita_verification"],
                        "suspicious_patterns": [
                            {"pattern": "recent_creation", "risk": "medium"},
                            {"pattern": "inactive_status", "risk": "high"}
                        ]
                    }
                },
                "MZ": {
                    "nuit": {
                        "min_confidence_score": 0.75,
                        "required_steps": ["format_check", "checksum_validation"],
                        "suspicious_patterns": [
                            {"pattern": "invalid_prefix", "risk": "high"},
                            {"pattern": "unregistered", "risk": "high"}
                        ]
                    },
                    "bi": {
                        "min_confidence_score": 0.75,
                        "required_steps": ["format_check", "checksum_validation"],
                        "validity_period": {"min_years": 5, "max_years": 10},
                        "suspicious_patterns": [
                            {"pattern": "invalid_format", "risk": "high"},
                            {"pattern": "expired_document", "risk": "medium"}
                        ]
                    }
                },
                "PT": {
                    "cc": {
                        "min_confidence_score": 0.85,
                        "required_steps": ["format_check", "checksum_validation", "doc_verification"],
                        "validity_period": {"min_years": 5, "max_years": 10},
                        "suspicious_patterns": [
                            {"pattern": "invalid_format", "risk": "high"},
                            {"pattern": "expired_document", "risk": "medium"},
                            {"pattern": "altered_document", "risk": "high"}
                        ]
                    },
                    "nif": {
                        "min_confidence_score": 0.85,
                        "required_steps": ["format_check", "checksum_validation"],
                        "suspicious_patterns": [
                            {"pattern": "invalid_prefix", "risk": "high"},
                            {"pattern": "unregistered", "risk": "high"}
                        ]
                    }
                }
            }
        
        try:
            # Carregar regras de arquivo
            import yaml
            with open(self.rules_path, 'r') as file:
                return yaml.safe_load(file)
        except Exception as e:
            logger.error(f"Erro ao carregar regras do arquivo: {str(e)}. Usando regras padrão.")
            return self._load_rules()
    
    def _load_model(self):
        """
        Carrega o modelo de ML para detecção de fraudes documentais.
        
        Returns:
            object: Modelo de ML carregado
        """
        if not self.model_path:
            # Modelo simulado simples
            class DummyModel:
                def predict_proba(self, features):
                    # Simula probabilidades baseado em features
                    import random
                    if features.get('confidence_score', 0) < 0.7:
                        return [random.uniform(0.6, 0.9)]
                    elif features.get('validation_result') is False:
                        return [random.uniform(0.7, 0.95)]
                    elif len(features.get('errors', [])) > 0:
                        return [random.uniform(0.5, 0.8)]
                    else:
                        return [random.uniform(0.1, 0.3)]
            
            return DummyModel()
        
        try:
            # Carregar modelo real
            import joblib
            return joblib.load(self.model_path)
        except Exception as e:
            logger.error(f"Erro ao carregar modelo: {str(e)}. Usando modelo simplificado.")
            return self._load_model()
    
    def on_consumer_start(self):
        """Ações ao iniciar o consumidor."""
        logger.info("Consumidor de validação documental iniciado. Processando eventos...")
    
    def on_consumer_stop(self):
        """Ações ao parar o consumidor."""
        self.executor.shutdown(wait=True)
        logger.info("Executor de threads encerrado.")
        
        # Log de estatísticas
        logger.info(f"Estatísticas de validação documental:")
        logger.info(f"Total de documentos processados: {self.validation_stats['total_documents']}")
        logger.info(f"Documentos válidos: {self.validation_stats['valid_documents']}")
        logger.info(f"Documentos inválidos: {self.validation_stats['invalid_documents']}")
        logger.info(f"Documentos suspeitos: {self.validation_stats['suspicious_documents']}")
        if self.validation_stats['total_documents'] > 0:
            valid_percentage = (self.validation_stats['valid_documents'] / self.validation_stats['total_documents']) * 100
            logger.info(f"Taxa de validade: {valid_percentage:.2f}%")
    
    def process_event(self, topic: str, event: Dict[str, Any]) -> ProcessingResult:
        """
        Processa um evento de validação documental.
        
        Args:
            topic: Tópico de onde o evento foi recebido
            event: Evento de validação documental
            
        Returns:
            ProcessingResult: Resultado do processamento
        """
        try:
            # Extrair informações básicas do evento
            metadata = event.get('metadata', {})
            document_type = event.get('document_type', 'unknown')
            document_number = event.get('document_number', 'unknown')
            country_code = event.get('country_code', 'unknown')
            validation_result = event.get('validation_result', False)
            confidence_score = event.get('confidence_score', 0.0)
            validation_steps = event.get('validation_steps', [])
            errors = event.get('errors', [])
            warnings = event.get('warnings', [])
            
            # Log de processamento
            logger.debug(f"Processando validação do documento {document_type}:{document_number} " +
                         f"(país: {country_code}, resultado: {validation_result})")
            
            # Atualizar estatísticas
            self.validation_stats['total_documents'] += 1
            
            if validation_result:
                self.validation_stats['valid_documents'] += 1
            else:
                self.validation_stats['invalid_documents'] += 1
            
            # Contar por tipo de documento
            doc_type_key = f"{country_code}:{document_type}"
            self.validation_stats['documents_by_type'][doc_type_key] = self.validation_stats['documents_by_type'].get(doc_type_key, 0) + 1
            
            # Contar por país
            self.validation_stats['documents_by_country'][country_code] = self.validation_stats['documents_by_country'].get(country_code, 0) + 1
            
            # Atualizar média de confiança
            current_avg = self.validation_stats['avg_confidence_score']
            count = self.validation_stats['total_documents']
            self.validation_stats['avg_confidence_score'] = ((current_avg * (count - 1)) + confidence_score) / count
            
            # Analisar o evento em busca de sinais de fraude
            fraud_signals, is_suspicious = self._analyze_fraud_signals(event)
            
            if is_suspicious:
                self.validation_stats['suspicious_documents'] += 1
                
                # Registrar sinais de fraude identificados
                for signal in fraud_signals:
                    signal_type = signal.get('type', 'unknown')
                    self.validation_stats['fraud_signals'][signal_type] = self.validation_stats['fraud_signals'].get(signal_type, 0) + 1
                
                # Se for uma tentativa suspeita, gerar alerta
                if fraud_signals:
                    self._generate_fraud_alert(event, fraud_signals)
            
            return ProcessingResult.success_result(
                f"Documento {document_type}:{document_number} processado com sucesso",
                data={
                    "document_type": document_type,
                    "document_number": document_number,
                    "country_code": country_code,
                    "validation_result": validation_result,
                    "confidence_score": confidence_score,
                    "fraud_signals": fraud_signals,
                    "is_suspicious": is_suspicious
                }
            )
            
        except KeyError as ke:
            error_msg = f"Erro ao processar evento: campo obrigatório não encontrado - {str(ke)}"
            logger.error(error_msg)
            return ProcessingResult.failure_result(error_msg, ke)
        
        except Exception as e:
            error_msg = f"Erro inesperado ao processar evento de validação documental: {str(e)}"
            logger.exception(error_msg)
            return ProcessingResult.failure_result(error_msg, e)
    
    def _analyze_fraud_signals(self, event: Dict[str, Any]) -> tuple:
        """
        Analisa o evento em busca de sinais de fraude.
        
        Args:
            event: Evento de validação documental
            
        Returns:
            tuple: (Lista de sinais de fraude, Flag indicando se é suspeito)
        """
        fraud_signals = []
        is_suspicious = False
        
        try:
            # Extrair informações básicas do evento
            document_type = event.get('document_type', 'unknown').lower()
            country_code = event.get('country_code', 'unknown').upper()
            validation_result = event.get('validation_result', False)
            confidence_score = event.get('confidence_score', 0.0)
            validation_steps = event.get('validation_steps', [])
            errors = event.get('errors', [])
            warnings = event.get('warnings', [])
            
            # Obter regras específicas para este tipo de documento/país
            rules = None
            if country_code in self.rules and document_type in self.rules[country_code]:
                rules = self.rules[country_code][document_type]
            else:
                rules = self.rules.get('default', {})
            
            # Verificar pontuação mínima de confiança
            min_confidence = rules.get('min_confidence_score', 0.7)
            if confidence_score < min_confidence:
                fraud_signals.append({
                    'type': 'low_confidence_score',
                    'description': f'Pontuação de confiança abaixo do mínimo exigido ({confidence_score} < {min_confidence})',
                    'risk_level': 'medium'
                })
                is_suspicious = True
            
            # Verificar etapas de validação obrigatórias
            required_steps = rules.get('required_steps', [])
            completed_steps = [step.get('name') for step in validation_steps]
            missing_steps = [step for step in required_steps if step not in completed_steps]
            
            if missing_steps:
                fraud_signals.append({
                    'type': 'missing_validation_steps',
                    'description': f'Etapas de validação obrigatórias não realizadas: {", ".join(missing_steps)}',
                    'missing_steps': missing_steps,
                    'risk_level': 'high'
                })
                is_suspicious = True
            
            # Verificar etapas de validação que falharam
            failed_steps = [step.get('name') for step in validation_steps if not step.get('result', True)]
            if failed_steps:
                fraud_signals.append({
                    'type': 'failed_validation_steps',
                    'description': f'Etapas de validação falhas: {", ".join(failed_steps)}',
                    'failed_steps': failed_steps,
                    'risk_level': 'medium'
                })
                is_suspicious = True
            
            # Verificar validade temporal (se aplicável)
            validity_period = rules.get('validity_period')
            if validity_period and 'document_issue_date' in event and 'document_expiry_date' in event:
                try:
                    issue_date = datetime.datetime.fromisoformat(event['document_issue_date'])
                    expiry_date = datetime.datetime.fromisoformat(event['document_expiry_date'])
                    current_date = datetime.datetime.now()
                    
                    # Verificar se o documento está expirado
                    if current_date > expiry_date:
                        fraud_signals.append({
                            'type': 'expired_document',
                            'description': f'Documento expirado em {expiry_date.strftime("%d/%m/%Y")}',
                            'risk_level': 'medium'
                        })
                        is_suspicious = True
                    
                    # Verificar período de validade incomum
                    validity_years = (expiry_date.year - issue_date.year)
                    min_years = validity_period.get('min_years', 0)
                    max_years = validity_period.get('max_years', 100)
                    
                    if validity_years < min_years or validity_years > max_years:
                        fraud_signals.append({
                            'type': 'unusual_validity_period',
                            'description': f'Período de validade incomum: {validity_years} anos (esperado entre {min_years} e {max_years})',
                            'risk_level': 'high'
                        })
                        is_suspicious = True
                except:
                    pass
            
            # Usar modelo ML para predição de fraude (se disponível)
            try:
                # Preparar features para o modelo
                features = {
                    'confidence_score': confidence_score,
                    'validation_result': validation_result,
                    'failed_steps_count': len(failed_steps),
                    'errors': errors,
                    'warnings': warnings,
                    'document_type': document_type,
                    'country_code': country_code
                }
                
                # Predição do modelo
                fraud_probability = self.model.predict_proba(features)[0]
                
                # Se probabilidade alta, adicionar sinal
                if fraud_probability > 0.7:
                    fraud_signals.append({
                        'type': 'ml_fraud_detection',
                        'description': f'Modelo de ML detectou alta probabilidade de fraude: {fraud_probability:.2f}',
                        'probability': fraud_probability,
                        'risk_level': 'high'
                    })
                    is_suspicious = True
            except Exception as model_error:
                logger.warning(f"Erro ao executar modelo ML: {str(model_error)}")
            
            # Verificar padrões suspeitos específicos para este documento
            suspicious_patterns = rules.get('suspicious_patterns', [])
            for pattern in suspicious_patterns:
                pattern_name = pattern.get('pattern')
                risk_level = pattern.get('risk', 'medium')
                
                # Verificar padrões comuns
                if pattern_name == 'invalid_format' and any('format' in error.lower() for error in errors):
                    fraud_signals.append({
                        'type': pattern_name,
                        'description': 'Formato do documento inválido',
                        'risk_level': risk_level
                    })
                    is_suspicious = True
                
                elif pattern_name == 'multiple_failures' and len(failed_steps) > 1:
                    fraud_signals.append({
                        'type': pattern_name,
                        'description': f'Múltiplas falhas de validação: {len(failed_steps)}',
                        'risk_level': risk_level
                    })
                    is_suspicious = True
                
                elif pattern_name == 'multiple_attempts' and 'attempt_number' in event and event['attempt_number'] > 3:
                    fraud_signals.append({
                        'type': pattern_name,
                        'description': f'Múltiplas tentativas de validação: {event["attempt_number"]}',
                        'risk_level': risk_level
                    })
                    is_suspicious = True
                
                # Outros padrões específicos podem ser implementados aqui
                
            return fraud_signals, is_suspicious
            
        except Exception as e:
            logger.error(f"Erro ao analisar sinais de fraude: {str(e)}")
            return [], False
    
    def _generate_fraud_alert(self, event: Dict[str, Any], fraud_signals: List[Dict[str, Any]]):
        """
        Gera um alerta de fraude baseado em sinais detectados.
        
        Args:
            event: Evento original
            fraud_signals: Sinais de fraude detectados
        """
        try:
            # Determinar o nível de severidade com base nos sinais
            severity_levels = [signal.get('risk_level', 'medium') for signal in fraud_signals]
            if 'high' in severity_levels:
                severity = 'high'
            elif 'medium' in severity_levels:
                severity = 'medium'
            else:
                severity = 'low'
            
            # Criar mensagem de alerta
            metadata = event.get('metadata', {})
            alert_message = (
                f"Alerta de fraude documental detectado: " +
                f"{event.get('document_type', 'unknown')}:{event.get('document_number', 'unknown')}"
            )
            
            # Detalhes do alerta
            alert_details = {
                'document_type': event.get('document_type', 'unknown'),
                'document_number': event.get('document_number', 'unknown'),
                'country_code': event.get('country_code', 'unknown'),
                'tenant_id': metadata.get('tenant_id', ''),
                'user_id': metadata.get('user_id', ''),
                'fraud_signals': fraud_signals,
                'validation_steps': event.get('validation_steps', []),
                'confidence_score': event.get('confidence_score', 0.0),
                'errors': event.get('errors', []),
                'warnings': event.get('warnings', []),
                'detection_timestamp': datetime.datetime.now().isoformat(),
                'original_event_id': metadata.get('event_id', ''),
                'severity': severity
            }
            
            # Publicar alerta (em implementação futura, integração com EventProducer)
            logger.warning(f"ALERTA DE FRAUDE: {alert_message}")
            logger.warning(f"Severidade: {severity.upper()}")
            logger.warning(f"Sinais detectados: {len(fraud_signals)}")
            
            # TODO: Integrar com EventProducer para publicar alerta
            # from infrastructure.fraud_detection.event_producers import EventProducer, AlertEvent, EventType, EventSeverity, EventMetadata
            # event_producer = EventProducer()
            # alert_metadata = EventMetadata(
            #     region_code=metadata.get('region_code', ''),
            #     tenant_id=metadata.get('tenant_id', ''),
            #     user_id=metadata.get('user_id', ''),
            #     source_module="document_validation_consumer"
            # )
            # alert_event = AlertEvent(
            #     metadata=alert_metadata,
            #     alert_id=str(uuid.uuid4()),
            #     alert_type="document_fraud",
            #     alert_severity=EventSeverity.HIGH if severity == 'high' else (EventSeverity.MEDIUM if severity == 'medium' else EventSeverity.LOW),
            #     source_events=[{'event_id': metadata.get('event_id', '')}],
            #     alert_message=alert_message,
            #     alert_details=alert_details
            # )
            # event_producer.produce_event(alert_event, EventType.ALERT)
            # event_producer.flush()
            
        except Exception as e:
            logger.error(f"Erro ao gerar alerta de fraude: {str(e)}")


# Função de exemplo para iniciar o consumidor
def start_document_validation_consumer(region_code=None):
    """
    Inicia o consumidor de validação documental.
    
    Args:
        region_code: Código da região para filtrar eventos
    """
    consumer = DocumentValidationConsumer(
        consumer_group_id=f"document_validation_consumer_{region_code or 'all'}",
        region_code=region_code
    )
    consumer.start()
    
    
if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Consumidor de eventos de validação documental")
    parser.add_argument("--region", help="Código da região (AO, BR, MZ, PT)", default=None)
    parser.add_argument("--config", help="Caminho para o arquivo de configuração", default=None)
    parser.add_argument("--model", help="Caminho para o modelo de ML", default=None)
    parser.add_argument("--rules", help="Caminho para regras de validação", default=None)
    
    args = parser.parse_args()
    
    consumer = DocumentValidationConsumer(
        region_code=args.region,
        config_file=args.config,
        model_path=args.model,
        rules_path=args.rules
    )
    
    try:
        consumer.start()
    except KeyboardInterrupt:
        logger.info("Consumidor interrompido pelo usuário.")
    except Exception as e:
        logger.error(f"Erro ao executar consumidor: {str(e)}")
    finally:
        consumer.stop()