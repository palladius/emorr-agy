# Track Spec: Telegram Channel Setup & Tragicomic Assets

## Overview
Establish Telegram channel status broadcasting capabilities for `emorr-agy`, document the channel configuration process, and generate a new tragicomic conductor asset to represent the project.

## Functional Requirements
1. **Telegram Channel Broadcasting**:
   - Add support in `emorr-agy` config to read a `TELEGRAM_CHANNEL_ID` from the environment.
   - Implement broadcasting routines to automatically publish host reboots, status notifications, and critical alerts directly to the configured Telegram channel.
2. **Tragicomic Conductor Asset**:
   - Generate a high-fidelity 3D cartoon illustration of the Italian orchestra conductor looking tragicomically sad/distressed, holding a baton in the Swiss countryside, gazing helplessly at the boring grey corporate office building ("E. Morricone A.g.").
   - Save this image in the repository assets under `doc/img/conductor_sad.png`.
3. **Documentation**:
   - Create `doc/telegram_setup.md` explaining:
     - Telegram channel creation steps.
     - Adding the Bot as an administrator.
     - Finding and setting `TELEGRAM_CHANNEL_ID` and `TELEGRAM_APITOKEN` in the `.env` configuration.
