# Kafka Topics Fix

## Problem

```
❌ Error from consumer: kafka server: The request attempted to perform an operation on an invalid topic
```

## Ursache

Kafka unterstützt **keine Wildcards** wie `blog.*` oder `twitchbot.*` in Consumer-Subscriptions.

## Lösung

### 1. Liste verfügbare Kafka Topics

```bash
kafka-topics --list --bootstrap-server localhost:9092
```

Oder mit Docker:
```bash
docker exec -it <kafka-container-name> kafka-topics --list --bootstrap-server localhost:9092
```

### 2. Kopiere die exakten Topic-Namen

Du solltest Topics sehen wie:
```
blog.posts
blog.comments
blog.categories
blog.tags
blog.media
twitchbot.streams
twitchbot.messages
twitchbot.viewers
twitchbot.clips
twitchbot.commands
link.links
link.clicks
```

### 3. Update `.env` mit exakten Namen

**❌ FALSCH (Wildcards):**
```env
KAFKA_TOPICS=blog.*,twitchbot.*,link.*
```

**✅ RICHTIG (Explizite Namen):**
```env
KAFKA_TOPICS=blog.posts,blog.comments,blog.categories,blog.tags,blog.media,twitchbot.streams,twitchbot.messages,twitchbot.viewers,twitchbot.clips,twitchbot.commands,link.links,link.clicks
```

### 4. Service neu starten

```bash
go run cmd/server/main.go
```

## Wichtig zu verstehen

- **Kafka Subscription**: Braucht exakte Topic-Namen (`blog.posts`)
- **SSE Event Filtering**: Unterstützt Wildcards (`blog.*`)

Beispiel:
```bash
# Kafka Consumer liest von: blog.posts, twitchbot.streams
# SSE Client filtert mit: event_types=blog.*

curl -N "http://localhost:8084/events?event_types=blog.*"
```

Der Client bekommt nur `blog.*` Events, auch wenn der Kafka Consumer mehrere Topics liest.

## Automatisches Script (Linux/Mac)

```bash
./scripts/list-kafka-topics.sh
```

## Wenn keine Topics existieren

Falls `kafka-topics --list` **leer** ist:
1. Starte die anderen Services (blog, twitchbot, link)
2. Die Services erstellen Topics automatisch beim ersten Event
3. Warte ein paar Sekunden
4. Liste Topics erneut

## Default Topics (falls nichts funktioniert)

Minimale Konfiguration zum Testen:
```env
# Nur Blog und Twitchbot
KAFKA_TOPICS=blog.posts,twitchbot.messages
```

## Support

Bei weiteren Problemen:
- Check logs: `go run cmd/server/main.go`
- Test Kafka Connection: `telnet localhost 9092`
- Check Redpanda UI: `http://localhost:8080` (falls aktiviert)
