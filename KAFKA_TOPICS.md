# Kafka Topics - ToxicToastGo Monorepo

Granulare Kafka Topics für Event-Driven Architecture. Jeder Use-Case hat einen eigenen Topic.

## Naming Convention
Format: `{service}.{domain}.{action}`

## Blog Service Topics

### Posts
- `blog.post.created` - Neuer Post erstellt (Draft)
- `blog.post.updated` - Post aktualisiert
- `blog.post.published` - Post veröffentlicht
- `blog.post.unpublished` - Post zurück zu Draft
- `blog.post.deleted` - Post gelöscht
- `blog.post.restored` - Post wiederhergestellt

### Categories
- `blog.category.created` - Neue Kategorie erstellt
- `blog.category.updated` - Kategorie aktualisiert
- `blog.category.deleted` - Kategorie gelöscht

### Tags
- `blog.tag.created` - Neuer Tag erstellt
- `blog.tag.updated` - Tag aktualisiert
- `blog.tag.deleted` - Tag gelöscht

### Comments
- `blog.comment.created` - Neuer Kommentar erstellt
- `blog.comment.approved` - Kommentar genehmigt
- `blog.comment.rejected` - Kommentar abgelehnt
- `blog.comment.deleted` - Kommentar gelöscht

### Media
- `blog.media.uploaded` - Medien-Datei hochgeladen
- `blog.media.deleted` - Medien-Datei gelöscht
- `blog.media.thumbnail.generated` - Thumbnail generiert

## Twitchbot Service Topics

### Streams
- `twitchbot.stream.started` - Stream gestartet
- `twitchbot.stream.ended` - Stream beendet
- `twitchbot.stream.updated` - Stream-Info aktualisiert

### Messages
- `twitchbot.message.received` - Chat-Nachricht empfangen
- `twitchbot.message.deleted` - Nachricht gelöscht
- `twitchbot.message.timeout` - User getimeoutet

### Viewers
- `twitchbot.viewer.joined` - Viewer beigetreten
- `twitchbot.viewer.left` - Viewer verlassen
- `twitchbot.viewer.banned` - Viewer gebannt
- `twitchbot.viewer.unbanned` - Viewer entbannt
- `twitchbot.viewer.mod.added` - Moderator hinzugefügt
- `twitchbot.viewer.mod.removed` - Moderator entfernt
- `twitchbot.viewer.vip.added` - VIP hinzugefügt
- `twitchbot.viewer.vip.removed` - VIP entfernt

### Clips
- `twitchbot.clip.created` - Clip erstellt
- `twitchbot.clip.updated` - Clip aktualisiert
- `twitchbot.clip.deleted` - Clip gelöscht

### Commands
- `twitchbot.command.created` - Neuer Command erstellt
- `twitchbot.command.updated` - Command aktualisiert
- `twitchbot.command.deleted` - Command gelöscht
- `twitchbot.command.executed` - Command ausgeführt

## Link Service Topics

### Links
- `link.created` - Neuer Shortlink erstellt
- `link.updated` - Link aktualisiert
- `link.deleted` - Link gelöscht
- `link.expired` - Link abgelaufen
- `link.activated` - Link aktiviert
- `link.deactivated` - Link deaktiviert

### Clicks
- `link.clicked` - Link geklickt
- `link.click.fraud.detected` - Verdächtiger Klick erkannt

## Notification Service Topics

### Channels
- `notification.channel.created` - Neuer Kanal erstellt
- `notification.channel.updated` - Kanal aktualisiert
- `notification.channel.deleted` - Kanal gelöscht
- `notification.channel.enabled` - Kanal aktiviert
- `notification.channel.disabled` - Kanal deaktiviert

### Groups
- `notification.group.created` - Neue Gruppe erstellt
- `notification.group.updated` - Gruppe aktualisiert
- `notification.group.deleted` - Gruppe gelöscht

### Notifications
- `notification.sent` - Benachrichtigung gesendet
- `notification.delivered` - Benachrichtigung zugestellt
- `notification.failed` - Zustellung fehlgeschlagen
- `notification.read` - Benachrichtigung gelesen

### Subscribers
- `notification.subscriber.added` - Subscriber hinzugefügt
- `notification.subscriber.removed` - Subscriber entfernt
- `notification.subscriber.confirmed` - Subscription bestätigt
- `notification.subscriber.unsubscribed` - Abgemeldet

## Webhook Service Topics

### Webhooks
- `webhook.created` - Neuer Webhook erstellt
- `webhook.updated` - Webhook aktualisiert
- `webhook.deleted` - Webhook gelöscht
- `webhook.enabled` - Webhook aktiviert
- `webhook.disabled` - Webhook deaktiviert

### Deliveries
- `webhook.delivered` - Webhook erfolgreich zugestellt
- `webhook.failed` - Webhook-Zustellung fehlgeschlagen
- `webhook.retry` - Webhook wird erneut versucht
- `webhook.expired` - Webhook-Zustellung abgelaufen

## Foodfolio Service Topics

### Categories
- `foodfolio.category.created` - Kategorie erstellt
- `foodfolio.category.updated` - Kategorie aktualisiert
- `foodfolio.category.deleted` - Kategorie gelöscht

### Companies
- `foodfolio.company.created` - Hersteller erstellt
- `foodfolio.company.updated` - Hersteller aktualisiert
- `foodfolio.company.deleted` - Hersteller gelöscht

### Items
- `foodfolio.item.created` - Artikel erstellt
- `foodfolio.item.updated` - Artikel aktualisiert
- `foodfolio.item.deleted` - Artikel gelöscht

### Item Variants
- `foodfolio.variant.created` - Variante erstellt
- `foodfolio.variant.updated` - Variante aktualisiert
- `foodfolio.variant.deleted` - Variante gelöscht
- `foodfolio.variant.stock.low` - Bestand niedrig
- `foodfolio.variant.stock.empty` - Bestand leer

### Item Details
- `foodfolio.detail.created` - Item Detail erstellt (Einkauf)
- `foodfolio.detail.opened` - Item geöffnet
- `foodfolio.detail.expired` - Item abgelaufen
- `foodfolio.detail.expiring.soon` - Item läuft bald ab
- `foodfolio.detail.consumed` - Item konsumiert/gelöscht
- `foodfolio.detail.moved` - Item umgelagert (Location geändert)
- `foodfolio.detail.frozen` - Item eingefroren
- `foodfolio.detail.thawed` - Item aufgetaut

### Locations
- `foodfolio.location.created` - Lagerort erstellt
- `foodfolio.location.updated` - Lagerort aktualisiert
- `foodfolio.location.deleted` - Lagerort gelöscht

### Warehouses
- `foodfolio.warehouse.created` - Geschäft erstellt
- `foodfolio.warehouse.updated` - Geschäft aktualisiert
- `foodfolio.warehouse.deleted` - Geschäft gelöscht

### Receipts
- `foodfolio.receipt.created` - Kassenbon erstellt
- `foodfolio.receipt.scanned` - Kassenbon gescannt
- `foodfolio.receipt.ocr.completed` - OCR abgeschlossen
- `foodfolio.receipt.items.matched` - Items mit Artikeln verknüpft
- `foodfolio.receipt.deleted` - Kassenbon gelöscht

### Shopping Lists
- `foodfolio.shoppinglist.created` - Einkaufsliste erstellt
- `foodfolio.shoppinglist.updated` - Einkaufsliste aktualisiert
- `foodfolio.shoppinglist.deleted` - Einkaufsliste gelöscht
- `foodfolio.shoppinglist.item.added` - Item zur Liste hinzugefügt
- `foodfolio.shoppinglist.item.removed` - Item von Liste entfernt
- `foodfolio.shoppinglist.item.purchased` - Item gekauft
- `foodfolio.shoppinglist.completed` - Alle Items gekauft

## SSE Service Topics

Das SSE-Service konsumiert alle oben genannten Topics und streamt sie an verbundene Clients.

### Meta-Events (vom SSE Service selbst)
- `sse.client.connected` - Client verbunden
- `sse.client.disconnected` - Client getrennt
- `sse.client.subscribed` - Client hat Filter aktualisiert

## Topic-Verwendung

### Produzenten
Jeder Service sollte Events für seine eigenen Aktionen publishen.

**Beispiel (Blog Service):**
```go
// Nach Post-Erstellung
kafkaProducer.PublishEvent("blog.post.created", eventData)

// Nach Post-Veröffentlichung
kafkaProducer.PublishEvent("blog.post.published", eventData)
```

### Konsumenten
- **SSE Service**: Konsumiert alle Topics und streamt an Clients
- **Notification Service**: Konsumiert relevante Topics für Benachrichtigungen
- **Webhook Service**: Konsumiert alle Topics für Webhook-Zustellungen

### Consumer Groups
- `sse-service-group` - SSE Service
- `notification-service-group` - Notification Service
- `webhook-service-group` - Webhook Service
- `analytics-service-group` - Zukünftige Analytics

## Event-Payload Struktur

Jedes Event sollte diese Struktur haben:

```json
{
  "id": "uuid",
  "type": "blog.post.created",
  "source": "blog-service",
  "timestamp": "2025-01-07T12:00:00Z",
  "data": {
    // Service-spezifische Daten
  }
}
```

## Partitionierung

Topics mit hohem Durchsatz sollten mehrere Partitionen haben:
- `twitchbot.message.received` - 10 Partitionen (nach stream_id)
- `link.clicked` - 5 Partitionen (nach link_id)
- Andere Topics - 3 Partitionen (Standard)

## Retention

- **Standard**: 7 Tage
- **Analytics**: 30 Tage (für spätere Auswertung)
- **Audit**: 90 Tage (für Compliance)
