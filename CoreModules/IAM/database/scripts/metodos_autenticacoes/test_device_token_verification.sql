-- Testes de Verificação de Autenticação por Dispositivos e Tokens

-- 1. Teste de Token de Software
SELECT device.verify_software_token(
    '123456',
    'secret123',
    interval '30 seconds'
) AS software_token_test;

-- 2. Teste de Token de Hardware
SELECT device.verify_hardware_token(
    'token123',
    'serial123',
    'Yubico'
) AS hardware_token_test;

-- 3. Teste de Certificado Digital
SELECT device.verify_digital_certificate(
    'cert123',
    'issuer123',
    CURRENT_TIMESTAMP + interval '1 year'
) AS digital_certificate_test;

-- 4. Teste de Cartão Inteligente
SELECT device.verify_smart_card(
    'card123',
    '1234',
    'JavaCard'
) AS smart_card_test;

-- 5. Teste de Token de Segurança
SELECT device.verify_security_token(
    'token123',
    'seed123',
    'HOTP'
) AS security_token_test;

-- 6. Teste de Token de Autenticação
SELECT device.verify_auth_token(
    'token123',
    'user123',
    'device123'
) AS auth_token_test;

-- 7. Teste de Chave USB
SELECT device.verify_usb_key(
    'key123',
    'serial123',
    'Yubico'
) AS usb_key_test;

-- 8. Teste de Cartão de Crédito
SELECT device.verify_credit_card(
    '4111111111111111',
    '123',
    CURRENT_DATE + interval '1 year'
) AS credit_card_test;

-- 9. Teste de Cartão de Débito
SELECT device.verify_debit_card(
    '4111111111111111',
    '1234',
    CURRENT_DATE + interval '1 year'
) AS debit_card_test;

-- 10. Teste de Cartão de Identificação
SELECT device.verify_id_card(
    'id123',
    'John Doe',
    CURRENT_DATE + interval '5 years'
) AS id_card_test;

-- 11. Teste de Cartão de Segurança
SELECT device.verify_security_card(
    'sec123',
    3,
    CURRENT_TIMESTAMP + interval '2 years'
) AS security_card_test;

-- 12. Teste de Chip NFC
SELECT device.verify_nfc_chip(
    'nfc123',
    'ISO14443',
    1024
) AS nfc_chip_test;

-- 13. Teste de Cartão RFID
SELECT device.verify_rfid_card(
    'rfid123',
    '13.56MHz',
    5.0
) AS rfid_card_test;

-- 14. Teste de Cartão MIFARE
SELECT device.verify_mifare_card(
    'mif123',
    1024,
    'High'
) AS mifare_card_test;

-- 15. Teste de Cartão de Proximidade
SELECT device.verify_proximity_card(
    'prox123',
    5.0,
    0.8
) AS proximity_card_test;

-- 16. Teste de Cartão Contactless
SELECT device.verify_contactless_card(
    'cont123',
    'ISO14443',
    1024
) AS contactless_card_test;

-- 17. Teste de Cartão Contact
SELECT device.verify_contact_card(
    'cont123',
    'ISO7816',
    1024
) AS contact_card_test;

-- 18. Teste de Cartão Dual Interface
SELECT device.verify_dual_interface_card(
    'dual123',
    ARRAY['ISO14443', 'ISO7816'],
    1024
) AS dual_interface_card_test;

-- 19. Teste de Cartão Biométrico
SELECT device.verify_biometric_security_card(
    'bio123',
    'Fingerprint',
    4
) AS biometric_card_test;

-- 20. Teste de Cartão Híbrido
SELECT device.verify_hybrid_security_card(
    'hyb123',
    ARRAY['ISO14443', 'ISO7816', 'NFC'],
    5
) AS hybrid_card_test;
