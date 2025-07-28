# Deploying La Ninna Hedgehog App to Render.com

This guide provides step-by-step instructions for deploying the La Ninna Hedgehog Rescue Center Management System to [Render.com](https://render.com).

## Prerequisites

Before you begin, make sure you have:

1. A [Render.com](https://render.com) account
2. Access to the La Ninna Hedgehog App repository
3. (Optional) A Cloudinary account for image upload functionality

## Deployment Steps

### 1. Fork or Clone the Repository

If you don't already have the code in your own Git repository:

```bash
git clone https://github.com/your-username/laninna-hedgehog-app.git
cd laninna-hedgehog-app
```

Push this to your own Git repository (GitHub, GitLab, etc.) that Render can access.

### 2. Deploy to Render using the Blueprint

The repository includes a `render.yaml` file that defines the infrastructure needed to run the application on Render.

1. Log in to your Render dashboard
2. Click on the "New" button and select "Blueprint"
3. Connect your Git repository where you've pushed the code
4. Render will detect the `render.yaml` file and show you the resources that will be created
5. Click "Apply" to start the deployment process

### 3. Configure Secrets

After the initial deployment, you'll need to set up the following secrets:

#### Required Secret:

- **JWT_SECRET**: Used for securing authentication tokens

To set this secret:

1. Go to your Render dashboard
2. Select the "laninna-hedgehog-app" web service
3. Go to the "Environment" tab
4. Find the JWT_SECRET variable
5. Click "Edit" and enter a secure random string
6. Click "Save Changes"

You can generate a secure random string using:

```bash
openssl rand -base64 32
```

#### Optional Secrets (for Image Upload):

If you want to enable image uploads for hedgehogs, you'll need to configure Cloudinary:

1. Sign up for a free Cloudinary account at https://cloudinary.com/
2. Get your Cloud Name, API Key, and API Secret from the dashboard
3. In your Render dashboard, set the following secrets:

   - **CLOUDINARY_CLOUD_NAME**: Your Cloudinary cloud name
   - **CLOUDINARY_API_KEY**: Your Cloudinary API key
   - **CLOUDINARY_API_SECRET**: Your Cloudinary API secret

### 4. Verify Deployment

After deployment is complete:

1. Click on the URL provided by Render to access your application
2. Log in with the default credentials: `admin` / `admin123`
3. Verify that you can access all features
4. If you configured Cloudinary, test the image upload functionality

### 5. Customize Your Deployment

You can customize your deployment by modifying the `render.yaml` file:

- Change the region
- Adjust the plan
- Modify environment variables
- Change the disk size

After making changes, commit and push them to your repository. Render will automatically redeploy your application if auto-deploy is enabled.

## Troubleshooting

### Application Not Starting

Check the logs in the Render dashboard for error messages. Common issues include:

- Missing environment variables
- Database initialization errors
- Build failures

### Image Upload Not Working

If image uploads aren't working:

1. Verify that you've set all three Cloudinary environment variables
2. Check that your Cloudinary account is active
3. Look for error messages in the application logs

### Database Issues

The application uses a persistent disk for the SQLite database. If you're experiencing database issues:

1. Check that the disk is properly mounted
2. Verify that the DB_PATH environment variable is set to `/data/laninna.db`
3. Ensure AUTO_MIGRATE is set to `true` for the first deployment

## Maintenance

### Updating the Application

To update your application:

1. Pull the latest changes from the main repository
2. Push the changes to your Git repository
3. Render will automatically redeploy if auto-deploy is enabled

### Backing Up the Database

To back up your SQLite database:

1. Go to the Render dashboard
2. Select the "laninna-hedgehog-app" web service
3. Go to the "Shell" tab
4. Run the following command to create a backup:

```bash
cp /data/laninna.db /tmp/laninna-backup-$(date +%Y%m%d).db
```

5. Download the backup file from the `/tmp` directory

## Additional Resources

- [Render Documentation](https://render.com/docs)
- [Cloudinary Documentation](https://cloudinary.com/documentation)
- [La Ninna Project Guidelines](PROJECT_GUIDELINES.md)