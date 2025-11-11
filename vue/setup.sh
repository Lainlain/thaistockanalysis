#!/bin/bash

echo "🚀 Thai Stock Analysis - Vue Admin Panel Setup"
echo "================================================"
echo ""

# Check if we're in the vue directory
if [ ! -f "package.json" ]; then
    echo "❌ Error: Please run this script from the vue/ directory"
    echo "   cd vue && ./setup.sh"
    exit 1
fi

echo "📦 Installing dependencies..."
npm install

if [ $? -ne 0 ]; then
    echo "❌ Failed to install dependencies"
    exit 1
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Start the Go backend (in a separate terminal):"
echo "   cd .. && go run cmd/server/main.go"
echo ""
echo "2. Start the Vue development server:"
echo "   npm run dev"
echo ""
echo "3. Open your browser:"
echo "   http://localhost:3000"
echo ""
echo "📚 For more information, see:"
echo "   - README.md (full documentation)"
echo "   - QUICKSTART.md (quick start guide)"
echo "   - PROJECT_COMPLETE.md (project overview)"
echo ""
echo "Happy managing! 📈"
