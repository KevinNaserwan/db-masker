-- Production Database Dump Sample
CREATE TABLE users (
    id INT PRIMARY KEY,
    full_name VARCHAR(255),
    email VARCHAR(255),
    phone_number VARCHAR(50),
    nik VARCHAR(50),
    home_address TEXT,
    password_hash VARCHAR(255),
    monthly_salary DECIMAL(15,2)
);

CREATE TABLE orders (
    id INT PRIMARY KEY,
    user_id INT,
    shipping_address TEXT,
    total_amount DECIMAL(15,2),
    credit_card_masked VARCHAR(50)
);

-- Real User PII Data
INSERT INTO `users` (`id`, `full_name`, `email`, `phone_number`, `nik`, `home_address`, `password_hash`, `monthly_salary`) VALUES
(1, 'Kevin Naserwan', 'kevin@nasercorp.com', '081298765432', '3271012345678901', 'Jl. Sudirman No. 45, Jakarta', '$2y$12$eX4mpL3H4shUnbr34k4bl3P4ssw0rd...', 25000000.00),
(2, 'Siti Nurhaliza', 'siti@singer.my', '081311223344', '3271029876543210', 'Jl. Diponegoro No. 12, Bandung', '$2y$12$4n0th3rSup3rS3cr3tH4sh...', 18500000.00);

-- Orders linked by Foreign Key user_id
INSERT INTO `orders` (`id`, `user_id`, `shipping_address`, `total_amount`, `credit_card_masked`) VALUES
(101, 1, 'Jl. Sudirman No. 45, Jakarta', 1500000.00, '4111-XXXX-XXXX-1111'),
(102, 1, 'Office Tower Lt. 14, Jakarta', 320000.00, '4111-XXXX-XXXX-1111'),
(103, 2, 'Jl. Diponegoro No. 12, Bandung', 750000.00, '5500-XXXX-XXXX-8899');
