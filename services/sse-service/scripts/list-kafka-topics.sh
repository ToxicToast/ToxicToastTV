#!/bin/bash

# List all available Kafka topics
# Usage: ./scripts/list-kafka-topics.sh [broker-address]

BROKER=${1:-localhost:19092}

echo "📋 Listing Kafka topics from: $BROKER"
echo ""

kafka-topics --list --bootstrap-server "$BROKER" 2>/dev/null

if [ $? -ne 0 ]; then
    echo "❌ Error: Could not connect to Kafka broker at $BROKER"
    echo ""
    echo "Troubleshooting:"
    echo "  1. Check if Kafka/Redpanda is running"
    echo "  2. Verify the broker address is correct"
    echo "  3. Try: docker ps | grep kafka"
    exit 1
fi

echo ""
echo "✅ Done! Copy the topics you need to your .env KAFKA_TOPICS"
echo ""
echo "Example:"
echo "KAFKA_TOPICS=blog.posts,blog.comments,twitchbot.streams,twitchbot.messages"
