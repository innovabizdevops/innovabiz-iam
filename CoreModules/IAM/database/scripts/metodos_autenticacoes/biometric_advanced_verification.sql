-- Funções de Verificação de Autenticação Biométrica Avançada

-- 1. Análise de DNA Rápida
CREATE OR REPLACE FUNCTION biometric_advanced.verify_rapid_dna_analysis(
    p_dna_data TEXT,
    p_quality_threshold FLOAT,
    p_match_threshold FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados do DNA
    IF p_dna_data IS NULL OR LENGTH(p_dna_data) < 2048 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_threshold < 0.0 OR p_quality_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold de correspondência
    IF p_match_threshold < 0.0 OR p_match_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise do DNA
    IF (p_dna_data->>'sequence_quality'::FLOAT < p_quality_threshold OR 
        p_dna_data->>'match_score'::FLOAT < p_match_threshold) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 2. Reconhecimento de Orelha 3D
CREATE OR REPLACE FUNCTION biometric_advanced.verify_ear_3d_recognition(
    p_ear_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da orelha
    IF p_ear_data IS NULL OR jsonb_typeof(p_ear_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise 3D da orelha
    IF (p_ear_data->>'shape_similarity'::FLOAT < p_threshold OR 
        p_ear_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 3. Leitura Térmica Facial Avançada
CREATE OR REPLACE FUNCTION biometric_advanced.verify_advanced_thermal_face(
    p_thermal_data JSONB,
    p_temperature_range FLOAT[],
    p_quality_threshold FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados térmicos
    IF p_thermal_data IS NULL OR jsonb_typeof(p_thermal_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar faixa de temperatura
    IF p_temperature_range IS NULL OR array_length(p_temperature_range, 1) != 2 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_threshold < 0.0 OR p_quality_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise térmica facial
    IF (p_thermal_data->>'temperature_range'::FLOAT[] != p_temperature_range OR 
        p_thermal_data->>'quality_score'::FLOAT < p_quality_threshold) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 4. Leitura de Impressão Palmar 3D
CREATE OR REPLACE FUNCTION biometric_advanced.verify_3d_palm_print(
    p_palm_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da palma
    IF p_palm_data IS NULL OR jsonb_typeof(p_palm_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise 3D da palma
    IF (p_palm_data->>'pattern_match'::FLOAT < p_threshold OR 
        p_palm_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 5. Multiespectral Biométrico Avançado
CREATE OR REPLACE FUNCTION biometric_advanced.verify_advanced_multispectral(
    p_multispectral_data JSONB[],
    p_thresholds FLOAT[],
    p_quality_threshold FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados multispectrais
    IF p_multispectral_data IS NULL OR array_length(p_multispectral_data, 1) < 2 THEN
        RETURN FALSE;
    END IF;

    -- Verificar thresholds
    IF p_thresholds IS NULL OR array_length(p_thresholds, 1) < 2 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_threshold < 0.0 OR p_quality_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise multispectral
    FOR i IN 1..array_length(p_multispectral_data, 1) LOOP
        IF p_multispectral_data[i]->>'match_score'::FLOAT < p_thresholds[i] THEN
            RETURN FALSE;
        END IF;
    END LOOP;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 6. Reconhecimento Facial 3D Avançado
CREATE OR REPLACE FUNCTION biometric_advanced.verify_advanced_3d_face(
    p_face_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados do rosto
    IF p_face_data IS NULL OR jsonb_typeof(p_face_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise facial 3D
    IF (p_face_data->>'depth_accuracy'::FLOAT < p_threshold OR 
        p_face_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 7. Reconhecimento Labial Avançado
CREATE OR REPLACE FUNCTION biometric_advanced.verify_advanced_lip_recognition(
    p_lip_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados labiais
    IF p_lip_data IS NULL OR jsonb_typeof(p_lip_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise labial
    IF (p_lip_data->>'movement_pattern'::FLOAT < p_threshold OR 
        p_lip_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 8. Análise de Odor Corporal Avançada
CREATE OR REPLACE FUNCTION biometric_advanced.verify_advanced_body_odor(
    p_odor_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados do odor
    IF p_odor_data IS NULL OR jsonb_typeof(p_odor_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise de odor
    IF (p_odor_data->>'chemical_profile'::FLOAT < p_threshold OR 
        p_odor_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 9. Pulsação Vascular Avançada
CREATE OR REPLACE FUNCTION biometric_advanced.verify_advanced_vascular_pulse(
    p_pulse_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da pulsação
    IF p_pulse_data IS NULL OR jsonb_typeof(p_pulse_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise vascular
    IF (p_pulse_data->>'pattern_match'::FLOAT < p_threshold OR 
        p_pulse_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 10. Análise de DNA Rápida e Multiespectral
CREATE OR REPLACE FUNCTION biometric_advanced.verify_rapid_multispectral_dna(
    p_dna_data JSONB,
    p_spectral_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados do DNA
    IF p_dna_data IS NULL OR jsonb_typeof(p_dna_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados espectrais
    IF p_spectral_data IS NULL OR jsonb_typeof(p_spectral_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_dna_data->>'sequence_quality'::FLOAT < p_quality_score OR 
        p_spectral_data->>'match_score'::FLOAT < p_threshold) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 11. Reconhecimento de Orelha e Face 3D
CREATE OR REPLACE FUNCTION biometric_advanced.verify_ear_face_3d(
    p_ear_data JSONB,
    p_face_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da orelha
    IF p_ear_data IS NULL OR jsonb_typeof(p_ear_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do rosto
    IF p_face_data IS NULL OR jsonb_typeof(p_face_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_ear_data->>'shape_similarity'::FLOAT < p_threshold OR 
        p_face_data->>'depth_accuracy'::FLOAT < p_threshold OR 
        p_ear_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_face_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 12. Leitura Térmica e Vascular
CREATE OR REPLACE FUNCTION biometric_advanced.verify_thermal_vascular(
    p_thermal_data JSONB,
    p_vascular_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados térmicos
    IF p_thermal_data IS NULL OR jsonb_typeof(p_thermal_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados vasculares
    IF p_vascular_data IS NULL OR jsonb_typeof(p_vascular_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_thermal_data->>'temperature_range'::FLOAT[] != p_vascular_data->>'temperature_range'::FLOAT[] OR 
        p_vascular_data->>'pattern_match'::FLOAT < p_threshold OR 
        p_thermal_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 13. Leitura de Impressão Palmar e Face 3D
CREATE OR REPLACE FUNCTION biometric_advanced.verify_palm_face_3d(
    p_palm_data JSONB,
    p_face_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da palma
    IF p_palm_data IS NULL OR jsonb_typeof(p_palm_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do rosto
    IF p_face_data IS NULL OR jsonb_typeof(p_face_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_palm_data->>'pattern_match'::FLOAT < p_threshold OR 
        p_face_data->>'depth_accuracy'::FLOAT < p_threshold OR 
        p_palm_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_face_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 14. Multiespectral e Odor
CREATE OR REPLACE FUNCTION biometric_advanced.verify_multispectral_odor(
    p_multispectral_data JSONB,
    p_odor_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados multispectrais
    IF p_multispectral_data IS NULL OR jsonb_typeof(p_multispectral_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do odor
    IF p_odor_data IS NULL OR jsonb_typeof(p_odor_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_multispectral_data->>'match_score'::FLOAT < p_threshold OR 
        p_odor_data->>'chemical_profile'::FLOAT < p_threshold OR 
        p_multispectral_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_odor_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 15. Reconhecimento Labial e Vascular
CREATE OR REPLACE FUNCTION biometric_advanced.verify_lip_vascular(
    p_lip_data JSONB,
    p_vascular_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados labiais
    IF p_lip_data IS NULL OR jsonb_typeof(p_lip_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados vasculares
    IF p_vascular_data IS NULL OR jsonb_typeof(p_vascular_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_lip_data->>'movement_pattern'::FLOAT < p_threshold OR 
        p_vascular_data->>'pattern_match'::FLOAT < p_threshold OR 
        p_lip_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_vascular_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 16. Análise de DNA Rápida e Odor
CREATE OR REPLACE FUNCTION biometric_advanced.verify_rapid_dna_odor(
    p_dna_data JSONB,
    p_odor_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados do DNA
    IF p_dna_data IS NULL OR jsonb_typeof(p_dna_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do odor
    IF p_odor_data IS NULL OR jsonb_typeof(p_odor_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_dna_data->>'sequence_quality'::FLOAT < p_quality_score OR 
        p_odor_data->>'chemical_profile'::FLOAT < p_threshold OR 
        p_dna_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_odor_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 17. Reconhecimento de Orelha e Odor
CREATE OR REPLACE FUNCTION biometric_advanced.verify_ear_odor(
    p_ear_data JSONB,
    p_odor_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da orelha
    IF p_ear_data IS NULL OR jsonb_typeof(p_ear_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do odor
    IF p_odor_data IS NULL OR jsonb_typeof(p_odor_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_ear_data->>'shape_similarity'::FLOAT < p_threshold OR 
        p_odor_data->>'chemical_profile'::FLOAT < p_threshold OR 
        p_ear_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_odor_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 18. Leitura Térmica e DNA
CREATE OR REPLACE FUNCTION biometric_advanced.verify_thermal_dna(
    p_thermal_data JSONB,
    p_dna_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados térmicos
    IF p_thermal_data IS NULL OR jsonb_typeof(p_thermal_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do DNA
    IF p_dna_data IS NULL OR jsonb_typeof(p_dna_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_thermal_data->>'temperature_range'::FLOAT[] != p_dna_data->>'temperature_range'::FLOAT[] OR 
        p_dna_data->>'sequence_quality'::FLOAT < p_quality_score OR 
        p_thermal_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 19. Leitura de Impressão Palmar e DNA
CREATE OR REPLACE FUNCTION biometric_advanced.verify_palm_dna(
    p_palm_data JSONB,
    p_dna_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados da palma
    IF p_palm_data IS NULL OR jsonb_typeof(p_palm_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do DNA
    IF p_dna_data IS NULL OR jsonb_typeof(p_dna_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_palm_data->>'pattern_match'::FLOAT < p_threshold OR 
        p_dna_data->>'sequence_quality'::FLOAT < p_quality_score OR 
        p_palm_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_dna_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 20. Reconhecimento Labial e DNA
CREATE OR REPLACE FUNCTION biometric_advanced.verify_lip_dna(
    p_lip_data JSONB,
    p_dna_data JSONB,
    p_threshold FLOAT,
    p_quality_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar dados labiais
    IF p_lip_data IS NULL OR jsonb_typeof(p_lip_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar dados do DNA
    IF p_dna_data IS NULL OR jsonb_typeof(p_dna_data) != 'object' THEN
        RETURN FALSE;
    END IF;

    -- Verificar threshold
    IF p_threshold < 0.0 OR p_threshold > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Verificar qualidade
    IF p_quality_score < 0.0 OR p_quality_score > 1.0 THEN
        RETURN FALSE;
    END IF;

    -- Análise combinada
    IF (p_lip_data->>'movement_pattern'::FLOAT < p_threshold OR 
        p_dna_data->>'sequence_quality'::FLOAT < p_quality_score OR 
        p_lip_data->>'quality_score'::FLOAT < p_quality_score OR 
        p_dna_data->>'quality_score'::FLOAT < p_quality_score) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
