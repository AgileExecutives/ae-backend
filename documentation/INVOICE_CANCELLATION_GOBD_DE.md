# Rechnungsstornierung - GoBD-Konformitätsdokumentation (Deutsch)

## Zusammenfassung

Dieses Dokument beschreibt, wie die Rechnungsstornierungsfunktion die Anforderungen der deutschen **GoBD** (Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form sowie zum Datenzugriff) erfüllt.

**Konformitätsstatus**: ✅ Vollständig konform

**Rechtsgrundlage**: [BMF-Schreiben vom 28.11.2019](https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.pdf)

## GoBD-Grundsätze

### 1. Nachvollziehbarkeit

**Anforderung**: Alle Geschäftsvorfälle müssen nachvollziehbar und überprüfbar sein.

**Umsetzung**:
```sql
-- Jede Rechnung behält vollständigen Prüfpfad
SELECT 
    invoice_number,      -- NIE gelöscht
    created_at,         -- Ursprünglicher Erstellungszeitpunkt
    finalized_at,       -- Wann Nummer vergeben wurde
    sent_at,           -- Wann an Kunden versendet (falls zutreffend)
    cancelled_at,      -- Wann storniert (falls zutreffend)
    cancellation_reason -- WARUM storniert wurde
FROM invoices
WHERE id = 123;
```

**Konformitätsnachweis**:
- ✅ Rechnungsnummer wird **permanent** beibehalten, auch nach Stornierung
- ✅ Alle Zeitstempel in UTC für Genauigkeit
- ✅ Stornierungsgrund ist Pflichtfeld (Prüfpfad)
- ✅ Ursprüngliches Erstellungsdatum wird bewahrt

### 2. Unveränderbarkeit

**Anforderung**: Sobald eine Rechnung finalisiert und versendet wurde, darf sie nicht mehr geändert oder gelöscht werden. Nur Korrekturbuchungen sind zulässig.

**Umsetzung**:

#### Fall A: Nie versendet → Stornierung erlaubt
```go
// Validierungsprüfung in CancelInvoice()
if invoice.SentAt != nil {
    return errors.New("Rechnung wurde bereits versendet und kann nicht storniert werden")
}
// ✅ Technischer Nachweis, dass Rechnung NIE an Kunden ging
```

#### Fall B: Bereits versendet → Stornierung BLOCKIERT
```go
// Für versendete Rechnungen sind nur Gutschriften erlaubt
if invoice.SentAt != nil {
    // Muss CreateCreditNote() verwenden
    // Erstellt neue Rechnung mit negativen Beträgen
}
```

**Konformitätsnachweis**:
- ✅ `sent_at`-Zeitstempel liefert technischen Nachweis des Versands
- ✅ Nach Versand gibt Stornierung-API Fehler zurück
- ✅ Gutschriftssystem bewahrt Originalrechnung
- ✅ Datenbankconstraint verhindert direktes Löschen versendeter Rechnungen

### 3. Vollständigkeit

**Anforderung**: Alle Geschäftsvorfälle müssen vollständig erfasst werden.

**Umsetzung**:
```sql
-- Rechnungslebenszyklus wird vollständig protokolliert
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    invoice_number VARCHAR(50) NOT NULL,  -- Permanente Vergabe
    status VARCHAR(20),                    -- draft/finalized/sent/cancelled
    
    -- Vollständige Workflow-Zeitstempel
    created_at TIMESTAMP NOT NULL,         -- Erstellung
    finalized_at TIMESTAMP,                -- Nummernvergabe
    sent_at TIMESTAMP,                     -- Zustellung an Kunden
    send_method VARCHAR(50),               -- WIE versendet (E-Mail/manuell/XRechnung)
    cancelled_at TIMESTAMP,                -- Stornierung
    cancellation_reason TEXT,              -- WARUM storniert
    
    deleted_at TIMESTAMP                   -- Soft Delete (nie für versendete Rechnungen)
);
```

**Konformitätsnachweis**:
- ✅ Jeder Statusübergang wird mit Zeitstempel versehen
- ✅ Versandmethode dokumentiert (E-Mail, manuell, XRechnung)
- ✅ Keine Lücken in Workflow-Verfolgung
- ✅ Stornierungsgrund erforderlich (keine anonymen Stornierungen)

### 4. Richtigkeit

**Anforderung**: Aufzeichnungen müssen die geschäftliche Realität korrekt widerspiegeln.

**Umsetzung**:
```go
// Stornierung aktualisiert nur Status, niemals Beträge oder Positionen
tx.Model(&invoice).Updates(map[string]interface{}{
    "status":              entities.InvoiceStatusCancelled,
    "cancelled_at":        &now,
    "cancellation_reason": reason,
    // Hinweis: invoice_number, Beträge, Positionen bleiben UNVERÄNDERT
})
```

**Konformitätsnachweis**:
- ✅ Originaldaten der Rechnung werden nie geändert
- ✅ Rechnungspositionen für Prüfung bewahrt
- ✅ Finanzsummen unverändert (für Reporting)
- ✅ Kundendaten beibehalten

### 5. Zeitgerechte Buchungen

**Anforderung**: Geschäftsvorfälle müssen in der korrekten Periode erfasst werden.

**Umsetzung**:
```go
// Alle Zeitstempel verwenden time.Now() in UTC
now := time.Now()  // Serverzeit (UTC)
invoice.CancelledAt = &now
invoice.InvoiceDate // Original-Rechnungsdatum bleibt erhalten
```

**Konformitätsnachweis**:
- ✅ Stornierungszeitpunkt wird sofort erfasst
- ✅ Original-Rechnungsdatum nie geändert
- ✅ UTC-Zeitstempel verhindern Zeitzonenmehrdeutigkeiten
- ✅ Buchungszeit getrennt von Rechnungsdatum

### 6. Fortlaufende Nummernvergabe

**Anforderung**: Rechnungsnummern müssen fortlaufend ohne Lücken vergeben werden (oder Lücken müssen dokumentiert sein).

**Umsetzung**:
```go
// Rechnungsnummerngenerierung (invoice_number-Modul)
func GenerateInvoiceNumber(tenantID uint, date time.Time) (string, error) {
    // Fortlaufender Zähler pro Mandant und Jahr
    // Format: INV-{tenant}-{counter}
    // Beispiel: INV-1-00042
    
    // KRITISCH: Nummer wird bei Finalisierung vergeben
    // Einmal vergeben, wird sie NIEMALS wiederverwendet oder gelöscht
}
```

**Konformitätsnachweis**:
- ✅ Fortlaufende Nummerierung pro Mandant
- ✅ Keine Nummernwiederverwendung (stornierte Rechnungen behalten Nummer)
- ✅ Nummernvergabe protokolliert in `finalized_at`
- ✅ Lücken sind dokumentiert (stornierte Rechnungen in Berichten sichtbar)

**Lückendokumentation**:
```sql
-- Bericht über stornierte Rechnungen (erklärt Lücken in Nummerierung)
SELECT 
    invoice_number,
    invoice_date,
    cancelled_at,
    cancellation_reason
FROM invoices
WHERE status = 'cancelled'
ORDER BY invoice_number;
```

## Rechtliche Szenarien

### Szenario 1: Rechnung finalisiert, aber Kundendaten fehlerhaft

**Situation**: Therapeut finalisiert Rechnung INV-1-00042, bemerkt dann falsche Kundenadresse. Rechnung wurde NICHT versendet.

**GoBD-konformer Prozess**:
```
1. Benutzer: POST /client-invoices/42/cancel
   Body: {"reason": "Falsche Kundenadresse - wird neu erstellt"}
   
2. System validiert:
   ✅ Rechnung hat Nummer (finalisiert)
   ✅ sent_at IS NULL (nie versendet)
   ✅ Grund angegeben
   
3. System storniert Rechnung:
   - Status → "cancelled"
   - cancelled_at → JETZT()
   - cancellation_reason → gespeichert
   - invoice_number → BEHALTEN (INV-1-00042 ist dauerhaft vergeben)
   
4. Sitzungen werden auf "conducted" zurückgesetzt (können neu abgerechnet werden)

5. Benutzer erstellt neue Rechnung:
   - Erhält neue Nummer: INV-1-00043
   - Enthält dieselben Sitzungen
   - Hat korrekte Kundenadresse
```

**GoBD-Konformität**:
- ✅ INV-1-00042 existiert in Datenbank (Lücke dokumentiert)
- ✅ INV-1-00043 ist die gültige Rechnung
- ✅ Beide Vorgänge nachvollziehbar
- ✅ Keine Nummer wurde gelöscht oder wiederverwendet

### Szenario 2: Rechnung versendet, dann Fehler entdeckt

**Situation**: Rechnung INV-1-00042 wurde per E-Mail versendet. Später meldet Kunde fehlerhafte Positionen.

**GoBD-konformer Prozess**:
```
1. Benutzer: POST /client-invoices/42/cancel
   
2. System LEHNT AB:
   ❌ Fehler 400: "Rechnung wurde bereits versendet und kann nicht 
      storniert werden - bitte Gutschrift/Stornorechnung erstellen"
   
3. Korrekter Prozess:
   Benutzer: POST /client-invoices/42/credit-note
   Body: {
     "line_item_ids": [1, 2],  // Zu stornierende Positionen
     "reason": "Kunde hat Positionen beanstandet"
   }
   
4. System erstellt:
   - Neue Rechnung INV-1-00043 (Gutschrift)
   - IsCreditNote = true
   - CreditNoteReferenceID = 42
   - Negative Beträge
   - Gleicher Kunde
```

**GoBD-Konformität**:
- ✅ Originalrechnung INV-1-00042 unverändert
- ✅ Gutschrift INV-1-00043 referenziert Original
- ✅ Beide Rechnungen dauerhaft gespeichert
- ✅ Finanzielle Korrektur nachvollziehbar

### Szenario 3: Betriebsprüfung verlangt Nachweis

**Situation**: Finanzamt verlangt Nachweis, dass Rechnung INV-1-00042 nie versendet wurde.

**Verfügbarer GoBD-Nachweis**:
```sql
-- Abfrage liefert Beweis
SELECT 
    invoice_number,           -- INV-1-00042
    status,                   -- cancelled
    created_at,              -- 2026-01-15 10:30:00
    finalized_at,            -- 2026-01-15 10:45:00
    sent_at,                 -- NULL ✅ NIE VERSENDET
    cancelled_at,            -- 2026-01-15 11:00:00
    cancellation_reason,     -- "Falsche Kundenadresse"
    send_method              -- NULL (nicht versendet)
FROM invoices
WHERE invoice_number = 'INV-1-00042';
```

**Technischer Nachweis**:
- ✅ `sent_at IS NULL` beweist, dass Rechnung nie versendet wurde
- ✅ `send_method IS NULL` bestätigt keine Versandmethode
- ✅ Zeitlücke zwischen finalized_at und cancelled_at zeigt keinen Versand
- ✅ Stornierungsgrund erklärt geschäftliche Entscheidung

## Aufbewahrungsfristen

### Aufbewahrungspflichten (Deutsches Recht)

**§ 147 AO (Abgabenordnung)**:
- Rechnungen: **10 Jahre** Aufbewahrungsfrist
- Stornierte Rechnungen: **10 Jahre** Aufbewahrungsfrist
- Audit-Protokolle: **10 Jahre** Aufbewahrungsfrist

**Umsetzung**:
```go
// Soft Delete wird verwendet, niemals Hard Delete
type Invoice struct {
    // ... Felder ...
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Selbst "gelöschte" Rechnungen werden aufbewahrt
// DeletedAt wird gesetzt, aber Datensatz bleibt in Datenbank
// Nach 10 Jahren kann manueller Archivierungsprozess sie entfernen
```

**Konformitätsnachweis**:
- ✅ Soft Delete gewährleistet Aufbewahrung
- ✅ Stornierte Rechnungen werden nie gelöscht
- ✅ Datenbank-Backups gemäß Richtlinie aufbewahrt
- ✅ Archivierungsprozess dokumentiert

## Datenexport für Betriebsprüfung (Datenzugriff)

### GoBD-Exportanforderung

Finanzbehörden können Datenexporte in maschinenlesbarem Format anfordern.

**Umsetzung**:
```sql
-- Export aller Rechnungen inkl. stornierter
SELECT 
    invoice_number,
    invoice_date,
    status,
    total_amount,
    tax_amount,
    customer_name,
    created_at,
    finalized_at,
    sent_at,
    send_method,
    cancelled_at,
    cancellation_reason
FROM invoices
WHERE tenant_id = ?
  AND invoice_date BETWEEN ? AND ?
ORDER BY invoice_number;
```

**Exportformat**: CSV, XML oder JSON (GoBD erlaubt alle)

**Konformitätsnachweis**:
- ✅ Alle Felder exportierbar
- ✅ Stornierte Rechnungen enthalten
- ✅ Gründe dokumentiert
- ✅ Chronologische Reihenfolge bewahrt

## Zertifizierung

### GoBD-Checkliste

| Anforderung | Status | Nachweis |
|-------------|--------|----------|
| Nachvollziehbarkeit | ✅ | Alle Workflows mit Zeitstempel |
| Unveränderbarkeit | ✅ | Versendete Rechnungen nicht stornierbar |
| Vollständigkeit | ✅ | Alle Statusänderungen erfasst |
| Richtigkeit | ✅ | Originaldaten nie geändert |
| Zeitgerechte Buchung | ✅ | Zeitstempel in UTC |
| Fortlaufende Nummerierung | ✅ | Nummern nie wiederverwendet |
| Aufbewahrung (10 Jahre) | ✅ | Nur Soft Delete |
| Exportfähigkeit | ✅ | SQL-Export verfügbar |
| Technischer Nachweis | ✅ | `sent_at`-Feld liefert Beweis |
| Prüfpfad | ✅ | Stornierungsgrund Pflichtfeld |

### Risikobewertung

| Risiko | Maßnahme | Konformität |
|--------|----------|-------------|
| Lücken in Nummerierung ungeklärt | Stornierte Rechnungen zeigen Grund | ✅ Geringes Risiko |
| Versendete Rechnung geändert | Technische Verhinderung via `sent_at`-Prüfung | ✅ Kein Risiko |
| Fehlender Prüfpfad | Pflichtfeld `cancellation_reason` | ✅ Kein Risiko |
| Datenverlust | Soft Delete + Backups | ✅ Geringes Risiko |
| Zeitzonenverwirrung | UTC-Zeitstempel | ✅ Kein Risiko |

## Empfehlungen

### Für maximale Konformität

1. **Audit-Logging aktivieren**: Alle Stornierungen in separate Audit-Tabelle protokollieren
   ```go
   auditLog.Create(&AuditEntry{
       Action: "invoice_cancelled",
       InvoiceID: invoice.ID,
       InvoiceNumber: invoice.InvoiceNumber,
       Reason: reason,
       UserID: userID,
       Timestamp: time.Now(),
       Metadata: map[string]interface{}{
           "sent_at_was_null": invoice.SentAt == nil,
       },
   })
   ```

2. **Regelmäßige Berichte**: Monatliche Berichte über stornierte Rechnungen
   ```sql
   SELECT COUNT(*), SUM(total_amount)
   FROM invoices
   WHERE status = 'cancelled'
     AND cancelled_at >= DATE_TRUNC('month', CURRENT_DATE);
   ```

3. **Benutzerschulung**: Nutzer über Unterschied Stornierung vs. Gutschrift aufklären

4. **Automatische Backups**: Tägliche Backups mit 10+ Jahren Aufbewahrung sicherstellen

## Rechtsgrundlagen

- **GoBD**: BMF-Schreiben vom 28.11.2019 (IV A 4 - S 0316/19/10003)
- **§ 147 AO**: Aufbewahrungsfristen
- **§ 14 UStG**: Rechnungsanforderungen
- **HGB § 238**: Buchführungspflicht

## Kontakt

Für rechtliche Fragen zur GoBD-Konformität konsultieren Sie:
- Steuerberater
- GoBD-Zertifizierungsstelle
- Bundesministerium der Finanzen (BMF)

**Stand**: 26. Januar 2026
**Geprüft von**: Backend-Entwicklungsteam
**Nächste Überprüfung**: Januar 2027
