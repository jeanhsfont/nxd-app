# Um único serviço NXD no Cloud Run

O deploy está configurado para **um único** serviço no Cloud Run: **`nxd`**.

- **URL do app:** `https://nxd-925156909645.us-central1.run.app`
- **Imagem:** `us-central1-docker.pkg.dev/slideflow-prod/nxd-repo/nxd:latest`
- **Workflow:** `.github/workflows/deploy-gcp.yml` (push em `main`)
- **Cloud Build:** `cloudbuild.yaml` (se usar trigger no GCP)

---

## Por que aparecem dois (ou mais) serviços?

Se você ainda vê **hubsystem-backend**, **hubsystem-frontend**, **hubsystem-nxd** ou outros no Cloud Run, são **serviços antigos** de configurações anteriores. Eles não são mais criados pelo repositório; o código atual sobe só o serviço **`nxd`**.

---

## Apagar os serviços antigos (deixar só o `nxd`)

Faça isso **antes** ou **depois** de dar o próximo deploy. O ideal: apagar os antigos e em seguida fazer um deploy para ter só o **nxd** no ar.

1. **Login e projeto**
   ```bash
   gcloud auth login
   gcloud config set project slideflow-prod
   ```

2. **Listar serviços na região**
   ```bash
   gcloud run services list --project=slideflow-prod --region=us-central1 --format="table(SERVICE,REGION)"
   ```

3. **Apagar cada serviço antigo** (só os que **não** são `nxd`)

   Apague **um por um** (confirme o nome na lista acima):

   ```bash
   gcloud run services delete hubsystem-backend --project=slideflow-prod --region=us-central1 --quiet
   gcloud run services delete hubsystem-frontend --project=slideflow-prod --region=us-central1 --quiet
   gcloud run services delete hubsystem-nxd --project=slideflow-prod --region=us-central1 --quiet
   ```

   Use só os nomes que existirem na sua lista. O serviço **`nxd`** não deve ser apagado — é o único que o repositório usa.

4. **Conferir**
   ```bash
   gcloud run services list --project=slideflow-prod --region=us-central1
   ```
   Deve aparecer apenas **nxd**.

---

## Depois de apagar os antigos

- Front e API ficam em uma única URL: `https://nxd-925156909645.us-central1.run.app`
- O build de produção do front (`web-app/.env.production`) já usa essa URL em `VITE_API_URL`
- CORS no backend (`main.go`) permite só essa origem (e localhost)

Se ainda tiver trigger do Cloud Build para outro arquivo (ex.: `cloudbuild.nxd.yaml`), desative ou apague esse trigger no GCP para não criar outro serviço por engano.
