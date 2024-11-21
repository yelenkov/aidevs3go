
### First, make sure you have:

1. Authenticated with GCP:
```sh
gcloud auth login
gcloud auth application-default login
```
2. Set your project ID:
```sh
gcloud config set project YOUR_PROJECT_ID
```
3. Enable the Secret Manager API in your project:
```sh
gcloud services enable secretmanager.googleapis.com
```
4. Then run these commands in your project root directory:
```sh
# go mod init <path to github repo>
go mod init github.com/dawidjelenkowski/aidevs3go 

# Tidy up dependencies
go mod tidy

# Verify modules
go mod verify

# Download dependencies
go mod download

# Dependencies
go get cloud.google.com/go/secretmanager/apiv1
go get github.com/joho/godotenv
```

1. Create a service account
```sh
# Format: gcloud iam service-accounts create NAME --display-name "DISPLAY_NAME"
gcloud iam service-accounts create secret-manager-sa \
    --display-name "Secret Manager Service Account"
```
2. Get your project ID
```sh
gcloud config get-value project
```
3. Grant the necessary IAM roles to the service account
```sh
# Set your project ID
export PROJECT_ID=$(gcloud config get-value project)

# Grant Secret Manager Admin role (can create and manage secrets)
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:secret-manager-sa@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/secretmanager.admin"
```

Alternatively, for read-only access, use secretmanager.secretAccessor:
```sh
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:secret-manager-sa@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
```
4. Create and download the service account key JSON
```sh
gcloud iam service-accounts keys create service-account-key.json \
    --iam-account=secret-manager-sa@$PROJECT_ID.iam.gserviceaccount.com
```
5. Set the environment variable for authentication:
```sh
export GOOGLE_APPLICATION_CREDENTIALS="$PWD/service-account-key.json"
```
### Optional: 
To verify the service account
```sh
# List service accounts
gcloud iam service-accounts list

# List keys for the service account
gcloud iam service-accounts keys list \
    --iam-account=secret-manager-sa@$PROJECT_ID.iam.gserviceaccount.com
```
To clean up when no longer needed
```sh
# Delete the service account
gcloud iam service-accounts delete secret-manager-sa@$PROJECT_ID.iam.gserviceaccount.com

# Or just delete a specific key
gcloud iam service-accounts keys delete KEY_ID \
    --iam-account=secret-manager-sa@$PROJECT_ID.iam.gserviceaccount.com
```