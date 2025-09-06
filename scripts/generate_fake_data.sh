#!/bin/bash

# Script to generate fake conversation data for development

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Generating fake conversation data...${NC}"

# Default database path
DB_PATH=${1:-"bonsai.db"}

# Run the Go script
go run scripts/generate_fake_data.go "$DB_PATH"

echo -e "${GREEN}✓ Fake data generation complete!${NC}"
echo -e "Database location: ${DB_PATH}"
echo -e "You can now test your application with the generated conversation data."