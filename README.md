# idol-api deploy branch

This branch contains only the files needed to deploy the production app.

On the server:

```bash
git checkout deploy
git pull --ff-only origin deploy
./backend/scripts/deploy-production.sh
```
