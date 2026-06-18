# E-Morragy Web (e-morragy-web) 🌐

Questo documento descrive il concetto, l'interfaccia e i requisiti per la plancia di controllo web responsive di **E. Morricone Ag** (`emorr-agy`).

---

## 💡 Il Concetto
**E-Morragy Web** è una "super slick" responsive web application (basata su **Node.js** e Tailwind CSS) che monitora e orchestra tutte le sessioni `tmux` e i vari agenti attivi su Cloud, AGI, e host locali/remoti.

Fornisce un pannello di controllo visivo immediato con una logica di prioritizzazione intelligente "Human-in-the-Loop" per gestire l'attenzione dell'utente senza distrazioni.

---

## 🎨 L'Interfaccia e i Codici Colore (Veloce e Deterministica) 🚦

Per garantire massime prestazioni e velocità di caricamento, l'interfaccia principale si basa su **dati deterministici nativi di tmux/processi** (zero overhead di AI durante il rendering iniziale).

Gli agenti/sessioni vengono visualizzati in tre stati principali chiaramente distinti:

### 1. 🟢 VERDE — [Waiting for Input] (Interazione Richiesta)
- **Stato**: L'agente è bloccato e sta aspettando che tu (l'utente) inserisca una risposta o una conferma.
- **Posizione**: Sempre posizionati **in alto (in cima alla lista)** per richiamare subito l'attenzione.
- **Funzionalità**: Mostra l'ultimo prompt/domanda della console. Puoi cliccare sul box verde, digitare la risposta in un campo testo ed **iniettare l'input direttamente nella sessione tmux sottostante** (usando `tmux send-keys`), sbloccando l'agente all'istante!

### 2. 🔴🟠 ROSSO / ARANCIONE — [Work in Progress] (Lasciali Lavorare!)
- **Stato**: L'agente sta elaborando, scrivendo codice, facendo ricerche o eseguendo task attivamente.
- **Azione**: *"Lasciali lavorare!"* Non c'è bisogno che tu faccia nulla. Il pannello mostra un indicatore di progresso o attività.

### 3. ⚫ GRIGIO — [Dead Sessions] (Sessioni Terinate)
- **Stato**: Sessioni inattive, finite o andate in crash.
- **Azione**: Visualizzate in fondo. Forniscono un pulsante rapido per **"risuscitare" (riavviare)** la sessione/agente con lo stesso comando originario.

---

## 🧠 Abbellimento: Il Servizio di Insight in Background (Asincrono, ogni 30s)

Come "abbellimento" opzionale, un servizio in background gira in modo asincrono ogni 30 secondi. Se funziona bene, altrimenti *"chissene"* — non deve assolutamente rallentare la visualizzazione deterministica principale.

Questo servizio esegue un'analisi leggera del contesto della console (tramite LLM in background) e fornisce:

1. **Summary dello Stato**: Un micro-riassunto testuale di cosa sta facendo effettivamente l'agente in quella sessione.
2. **Priorità di Risposta**:
   - 🚨 *High Priority:* "Rispondimi quanto prima, stiamo aspettando te!"
   - ☕ *Low Priority:* "Chissene, non è urgente rispondere adesso."
3. **Suggerimento Risposte / Contesto**:
   - Propone 2 o 3 risposte preconfezionate rapide basate sull'ultima domanda dell'agente.
   - Mostra un estratto sintetico del contesto corrente dell'agente per aiutarti a rispondere senza dover rileggere tutto lo storico.
