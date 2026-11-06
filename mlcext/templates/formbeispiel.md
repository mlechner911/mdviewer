# {{title}}

<!--! **So arbeiten Sie mit diesem Formblatt:** Füllen Sie die Felder am Bildschirm aus und speichern Sie das PDF. Felder mit einem Ereignis-Bezug kann ein angebundenes System wie der Sensor Monitor vorfüllen. Dieser Hinweis ist aufklappbar und wird nicht gedruckt. -->

> [!NOTE]
> Beispiel für alle Feldarten der MLC-Erweiterung. Was hier steht, gehört zu {{doc_id}}, Fassung {{version}}. Nicht Benötigtes einfach löschen. Eine neue Seite beginnt mit `<div class="page-break"></div>` auf einer eigenen Zeile.

## 1. Vorausgefüllt aus dem Ereignis

Diese Feldnamen füllt ein angebundenes System wie der Sensor Monitor, wenn das Formblatt an einem Ereignis geöffnet wird. Ohne ein solches System sind es gewöhnliche Felder.

* **Ereignis:** [@#ereignis ____________]
* **Beginn:** [@#entdeckt_datum TT.MM.JJJJ] um [@#entdeckt_uhrzeit HH:MM] Uhr
* **Ort:** [@#ort ______________________________] <!--* Zone › Gerät › Probe -->
* **Betroffen:** [@#betroffen ______________________________]

**Vorfall:**

[@#vorfall ______________________________________________________________________
______________________________________________________________________
______________________________________________________________________]

## 2. Textfelder

* **Kurzes Feld:** [@#kuerzel ________]
* **Langes Feld:** [@#bemerkung ______________________________________________] <!--* Der Tooltip steht direkt nach dem Feld und wird nicht gedruckt. -->
* **Mit Vorgabewert:** [@#raum ____Kühlraum A____]
* **Ohne Namen:** [@____________] <!--* Ohne #name vergibt mlcdoc einen Namen; für Auswertungen besser benennen. -->

## 3. Datum, Uhrzeit, Zahlen

* **Datum:** [@#datum TT.MM.JJJJ]
* **Datum mit Vorgabe:** [@#stichtag TT.MM.JJJJ=01\.01\.2027]
* **Uhrzeit:** [@#uhrzeit HH:MM]
* **Temperatur:** [@#temperatur ______] °C

## 4. Ankreuzfelder

* [@#geprueft] Sichtprüfung durchgeführt
* [@#abweichung] Abweichung festgestellt
* [@#meldung] Meldung an QS erfolgt

## 5. Tabelle mit Feldern

| Maßnahme | Verantwortlich | Frist | Erledigt |
| :------------------------------------------------ | :----------------------- | :---------------- | :-------------: |
| [@#massnahme_01 ______________________________] | [@#verantwortlich_01 ______________] | [@#frist_01 TT.MM.JJJJ] | [@#erledigt_01] |
| [@#massnahme_02 ______________________________] | [@#verantwortlich_02 ______________] | [@#frist_02 TT.MM.JJJJ] | [@#erledigt_02] |

## 6. Mehrzeilige Notiz

[@#notiz ______________________________________________________________________
______________________________________________________________________]


## 7. Unterschrift

* **Name:** [@#name ________________________] **Datum:** [@#datum_unterschrift TT.MM.JJJJ]
* **Unterschrift:** ________________________
