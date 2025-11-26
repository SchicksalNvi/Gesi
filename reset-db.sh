#!/bin/bash

# Reset database script

echo "🗑️  Resetting database..."
echo ""

# Backup old database
if [ -f "data/cesi.db" ]; then
    timestamp=$(date +%Y%m%d_%H%M%S)
    echo "📦 Backing up old database to data/cesi.db.backup_$timestamp"
    cp data/cesi.db "data/cesi.db.backup_$timestamp"
fi

# Remove database files
echo "🗑️  Removing database files..."
rm -f data/cesi.db*

echo "✅ Database reset complete!"
echo ""
echo "You can now start the application with:"
echo "  ./start-frontend.sh"
