#!/bin/bash
set -e

echo "🎭 Setting up Playwright E2E tests..."

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install Node.js first."
    exit 1
fi

# Install dependencies
echo "📦 Installing npm dependencies..."
npm install

# Install Playwright browsers
echo "🌐 Installing Playwright browsers..."
npm run playwright:install

echo "✅ Setup complete!"
echo ""
echo "To run tests:"
echo "  npm run test:e2e          # Run all tests (headless)"
echo "  npm run test:e2e:ui       # Run with UI mode"
echo "  npm run test:e2e:headed   # Run in headed mode"
echo "  npm run test:e2e:debug    # Debug mode"
