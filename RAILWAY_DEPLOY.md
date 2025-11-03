# Railway Deployment Guide

## How PORT Works

**Important:** You don't need to set PORT in your `.env` file! Railway automatically provides the `PORT` environment variable when your service is deployed. 

- **On Railway:** Railway assigns a dynamic port and sets it as the `PORT` environment variable automatically
- **Locally:** If `PORT` is not set, the app falls back to port `8000` (as configured in `main.go`)

This is why your code uses:
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8000" // Default for local development
}
```

## Deployment Steps

### Method 1: Using Railway CLI (Recommended)

1. **Install Railway CLI:**
   ```bash
   npm install -g @railway/cli
   # OR using PowerShell on Windows
   iwr https://railway.app/install.sh | iex
   ```

2. **Login to Railway:**
   ```bash
   railway login
   ```

3. **Initialize Railway in your project:**
   ```bash
   railway init
   ```
   - This will create a new project or link to an existing one
   - Select "New Project" if this is your first time

4. **Add Database Service (PostgreSQL):**
   ```bash
   railway add postgres
   ```
   - This creates a PostgreSQL database
   - Railway automatically sets the `DATABASE_URL` environment variable

5. **Add Environment Variables:**
   ```bash
   railway variables set CLOUDINARY_CLOUD_NAME=your_cloud_name
   railway variables set CLOUDINARY_API_KEY=your_api_key
   railway variables set CLOUDINARY_API_SECRET=your_api_secret
   ```
   
   Or set them via the Railway dashboard:
   - Go to your project → Service → Variables
   - Add each variable

6. **Deploy:**
   ```bash
   railway up
   ```
   - This will build your Docker image and deploy to Railway

7. **Get Public URL:**
   ```bash
   railway domain
   ```
   - Or go to Railway dashboard → Settings → Generate Domain

### Method 2: Using GitHub Integration (Alternative)

1. **Push your code to GitHub** (if not already):
   ```bash
   git add .
   git commit -m "Prepare for Railway deployment"
   git push origin main
   ```

2. **Go to Railway Dashboard:**
   - Visit https://railway.app
   - Login with GitHub

3. **Create New Project:**
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository

4. **Add Services:**
   - Railway will detect your `Dockerfile` and `railway.json`
   - Click "Add Service" → "Database" → "Add PostgreSQL"
   - Railway will automatically link it and set `DATABASE_URL`

5. **Configure Environment Variables:**
   - Go to your service → Variables
   - Add:
     - `CLOUDINARY_CLOUD_NAME`
     - `CLOUDINARY_API_KEY`
     - `CLOUDINARY_API_SECRET`

6. **Deploy:**
   - Railway will automatically build and deploy
   - Go to Settings → Generate Domain to get your public URL

## Verification

After deployment, check:

1. **Railway automatically sets:**
   - `PORT` - The port your app should listen on (don't set this manually!)
   - `DATABASE_URL` - Connection string for PostgreSQL (if you added PostgreSQL service)
   - `RAILWAY_ENVIRONMENT` - The environment name

2. **Test your deployment:**
   ```bash
   curl https://your-app.railway.app/products
   ```

3. **Check logs:**
   ```bash
   railway logs
   ```
   Or view in Railway dashboard → Deployments → View Logs

## Important Notes

- **PORT is automatic:** Never manually set `PORT` in Railway variables
- **DATABASE_URL is automatic:** Railway sets this when you add a PostgreSQL service
- **Your app listens on 0.0.0.0:** This is correct - Railway routes traffic to your app
- **CORS is configured:** The CORS fix allows external access to your API

## Troubleshooting

### If deployment fails:
1. Check build logs: `railway logs` or view in dashboard
2. Verify Dockerfile is correct
3. Ensure all dependencies are in `go.mod`

### If CORS errors persist:
1. Make sure you've pushed the updated `main.go` with the CORS fix
2. Verify the deployment includes the latest code
3. Check browser console for specific CORS error messages

### If database connection fails:
1. Verify PostgreSQL service is added: `railway status`
2. Check `DATABASE_URL` is set: `railway variables`
3. Ensure database is provisioned and running

## Next Steps

Once deployed:
1. Share your Railway public URL with testers
2. Update your frontend to use the Railway URL instead of localhost
3. Monitor usage in Railway dashboard
4. Set up custom domain (optional) in Railway settings

