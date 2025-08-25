-- Create wallet_address table
CREATE TABLE IF NOT EXISTS wallet_address (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    address VARCHAR(255) NOT NULL,
    type address_type NOT NULL,
    standard address_standard
);

-- Create unique index on address
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_address ON wallet_address (address);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_wallet_address_type ON wallet_address (type);
CREATE INDEX IF NOT EXISTS idx_wallet_address_standard ON wallet_address (standard);
CREATE INDEX IF NOT EXISTS idx_wallet_address_created_at ON wallet_address (created_at);

-- Create enum types if they don't exist
DO $$ BEGIN
    CREATE TYPE address_type AS ENUM (
        'evm',
        'btc',
        'sol',
        'aptos',
        'tron'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE address_standard AS ENUM (
        'erc20',
        'erc721',
        'erc1155',
        'native',
        'spl',
        'trc20',
        'trc721'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Add comments for documentation
COMMENT ON TABLE wallet_address IS 'Stores wallet addresses for different blockchain networks';
COMMENT ON COLUMN wallet_address.address IS 'The wallet address string';
COMMENT ON COLUMN wallet_address.type IS 'The blockchain network type (evm, bitcoin, solana, tron)';
COMMENT ON COLUMN wallet_address.standard IS 'The token standard (erc20, erc721, etc.)';

-- Insert sample data
INSERT INTO wallet_address (address, type, standard) VALUES
('TAWdqnuYCNU3dKsi7pR8d7sDkx1Evb2giV', 'tron', 'trc20'),
('TT1j2adMBb6bF2K8C2LX1QkkmSXHjiaAfw', 'tron', 'trc20');