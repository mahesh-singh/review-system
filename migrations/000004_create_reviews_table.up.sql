-- Merged Reviews table (includes all comment and reviewer info, including room type)
CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    hotel_review_id BIGINT NOT NULL UNIQUE,
    hotel_id BIGINT NOT NULL,
    provider_id INTEGER NOT NULL,
    
    -- Review Details
    rating DECIMAL(3,1) NOT NULL,
    check_in_month_year VARCHAR(50),
    encrypted_review_data TEXT,
    formatted_rating VARCHAR(10),
    formatted_review_date VARCHAR(50),
    rating_text VARCHAR(50),
    responder_name VARCHAR(255),
    response_date_text VARCHAR(50),
    response_translate_source VARCHAR(10),
    review_comments TEXT,
    review_negatives TEXT,
    review_positives TEXT,
    review_provider_logo TEXT,
    review_provider_text VARCHAR(100),
    review_title VARCHAR(500),
    translate_source VARCHAR(10),
    translate_target VARCHAR(10),
    review_date TIMESTAMP,
    original_title VARCHAR(500),
    original_comment TEXT,
    formatted_response_date VARCHAR(50),
    is_show_review_response BOOLEAN DEFAULT FALSE,
    
    -- Reviewer Info 
    reviewer_country_name VARCHAR(100),
    reviewer_display_name VARCHAR(50),
    reviewer_flag_name VARCHAR(10),
    reviewer_group_name VARCHAR(100),
    reviewer_room_type_name VARCHAR(200), -- room_types 
    reviewer_country_id INTEGER,
    reviewer_length_of_stay INTEGER,
    reviewer_group_id INTEGER,
    reviewer_review_count INTEGER DEFAULT 0,
    reviewer_is_expert BOOLEAN DEFAULT FALSE,
    reviewer_show_global_icon BOOLEAN DEFAULT FALSE,
    reviewer_show_review_count BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (hotel_id) REFERENCES hotels(hotel_id) ON DELETE CASCADE,
    FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE RESTRICT,
    FOREIGN KEY (reviewer_country_id) REFERENCES countries(id) ON DELETE SET NULL,
    FOREIGN KEY (reviewer_group_id) REFERENCES review_groups(id) ON DELETE SET NULL
);