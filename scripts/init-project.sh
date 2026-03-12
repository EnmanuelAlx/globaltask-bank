#!/bin/bash

# Colors for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Checking environment...${NC}"

# Check .env
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env from .env.example...${NC}"
    cp .env.example .env
fi

# Check Supabase CLI
if ! command -v supabase >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Supabase CLI not found.${NC}"
    read -p "Do you want to install it now via NPM? (y/n) " yn
    if [ "$yn" = "y" ]; then
        echo -e "${BLUE}Installing Supabase CLI...${NC}"
        npm install -g supabase
    else
        echo -e "${RED}❌ Aborting. Please install Supabase CLI manually: https://supabase.com/docs/guides/cli${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}🚀 Starting Supabase services...${NC}"
# Use --no-watch to prevent blocking
supabase start --workdir infra

echo -e "${BLUE}💾 Running migrations...${NC}"
supabase db reset --workdir infra

echo -e "${BLUE}👤 Creating default developer user...${NC}"
EMAIL="admin@globaltask.com"
PASS="password123"

# Extract service_role key and API URL for local Supabase
STATUS_JSON=$(supabase status --workdir infra -o json)
SERVICE_ROLE_KEY=$(echo "$STATUS_JSON" | grep '"SERVICE_ROLE_KEY":' | sed -E 's/.*"SERVICE_ROLE_KEY": "(.*)",?/\1/' | tr -d '", ')
ANON_KEY=$(echo "$STATUS_JSON" | grep '"ANON_KEY":' | sed -E 's/.*"ANON_KEY": "(.*)",?/\1/' | tr -d '", ')
API_URL=$(echo "$STATUS_JSON" | grep '"API_URL":' | sed -E 's/.*"API_URL": "(.*)",?/\1/' | tr -d '", ')
DB_URL=$(echo "$STATUS_JSON" | grep '"DB_URL":' | sed -E 's/.*"DB_URL": "(.*)",?/\1/' | tr -d '", ')

# Create user via Supabase API
USER_RESPONSE=$(curl -s -X POST "$API_URL/auth/v1/admin/users" \
  -H "Authorization: Bearer $SERVICE_ROLE_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASS\",
    \"email_confirm\": true,
    \"user_metadata\": {
      \"full_name\": \"Admin\",
      \"role\": \"ADMIN\",
      \"identity_document\": \"MTIzNDU2Nzg5\",
      \"identity_document_bidx\": \"MTIzNDU2Nzg5\",
      \"country_id\": \"1\"
    }
  }")

# Extract user ID from response
USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$USER_ID" ]; then
  echo -e "${GREEN}✅ User created: $USER_ID${NC}"
else
  echo -e "${RED}❌ Failed to create user${NC}"
  echo "$USER_RESPONSE"
fi

echo -e "\n${GREEN}✨ Initialization Complete!${NC}"
echo -e "${BLUE}--------------------------------------------------${NC}"

echo -e "${GREEN}🔑 Supabase Info:${NC}"
echo -e "   ${YELLOW}Supabase URL:${NC} $API_URL"
echo -e "   ${YELLOW}Database URL:${NC} $DB_URL"
echo -e "   ${YELLOW}JWKS URL:${NC}     $API_URL/auth/v1/.well-known/jwks.json"
echo -e "   ${YELLOW}Anon Key:${NC}     $ANON_KEY"
echo -e "   ${YELLOW}Service Role Key:${NC} $SERVICE_ROLE_KEY"
echo -e ""
echo -e "${GREEN}👤 User Credentials:${NC}"
echo -e "   ${YELLOW}Email:${NC}    $EMAIL"
echo -e "   ${YELLOW}Password:${NC} $PASS"
echo -e "${BLUE}--------------------------------------------------${NC}"

echo -e "${YELLOW}👉 Next Step: Modify .env file and run '${NC}${GREEN}make up${NC}${YELLOW}' to build and start the containers.${NC}"
