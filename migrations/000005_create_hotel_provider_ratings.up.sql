CREATE TABLE IF NOT EXISTS hotel_provider_ratings (
    id BIGSERIAL PRIMARY KEY,
    hotel_id BIGINT NOT NULL,
    provider_id INTEGER NOT NULL,
    provider_name VARCHAR(100) NOT NULL,
    overall_score DECIMAL(3,1) NOT NULL,
    review_count INTEGER NOT NULL DEFAULT 0,
    
    -- Detailed Grades 
    cleanliness DECIMAL(3,1),
    facilities DECIMAL(3,1),
    location DECIMAL(3,1),
    room_comfort_quality DECIMAL(3,1),
    service DECIMAL(3,1),
    value_for_money DECIMAL(3,1),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (hotel_id) REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE RESTRICT,
    UNIQUE(hotel_id, provider_id)
);