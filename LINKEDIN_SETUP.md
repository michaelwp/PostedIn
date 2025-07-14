# LinkedIn API Setup Guide

This guide walks you through setting up LinkedIn API integration for the PostedIn scheduler.

## Prerequisites

1. A LinkedIn account
2. Access to LinkedIn Developer Portal

## Step 1: Create LinkedIn App

1. Go to [LinkedIn Developer Portal](https://developer.linkedin.com/)
2. Click "Create App"
3. Fill in the required information:
   - **App name**: Your app name (e.g., "PostedIn Scheduler")
   - **LinkedIn Page**: Associate with your LinkedIn company page (or create one)
   - **App logo**: Upload a logo (optional)
   - **Legal agreement**: Accept the terms

## Step 2: Configure App Permissions

1. In your app dashboard, go to the "Auth" tab
2. Add the following OAuth 2.0 scopes:
   - `openid`
   - `profile` 
   - `w_member_social` (for posting content)

3. Add OAuth 2.0 redirect URLs:
   - `http://localhost:8080/callback` (for local development)

## Step 3: Get API Credentials

1. In the "Auth" tab, note down:
   - **Client ID**
   - **Client Secret**

## Step 4: Configure PostedIn

1. Run the application for the first time:
   ```bash
   make run
   ```

2. This will create a `config.json` file with default values:
   ```json
   {
     "linkedin": {
       "client_id": "",
       "client_secret": "",
       "redirect_url": "http://localhost:8080/callback"
     },
     "storage": {
       "posts_file": "posts.json",
       "token_file": "linkedin_token.json"
     },
     "timezone": {
       "location": "Asia/Bangkok",
       "offset": "+07:00"
     }
   }
   ```

3. Edit `config.json` and fill in your LinkedIn credentials:
   ```json
   {
     "linkedin": {
       "client_id": "your_client_id_here",
       "client_secret": "your_client_secret_here",
       "redirect_url": "http://localhost:8080/callback"
     },
     "storage": {
       "posts_file": "posts.json",
       "token_file": "linkedin_token.json"
     },
     "timezone": {
       "location": "Asia/Bangkok",
       "offset": "+07:00"
     }
   }
   ```

## Step 5: Start the Callback Server (Optional but Recommended)

For a better authentication experience, start the callback server:

```bash
# In a separate terminal
make run-callback
# or
make build-callback && ./bin/callback-server
```

This provides:
- A user-friendly web interface for authentication
- Better error handling and feedback
- Automatic token saving
- Graceful shutdown handling

The server will run on `http://localhost:8080` by default.

## Step 6: Authenticate with LinkedIn

### Option A: Using the Callback Server (Recommended)

1. Make sure the callback server is running (Step 5)
2. Open `http://localhost:8080` in your browser
3. Click "Authenticate with LinkedIn"
4. Log in to LinkedIn and authorize the application
5. You'll see a success page with confirmation

### Option B: Using the CLI Authentication

1. Run the main application:
   ```bash
   make run
   ```

2. Select option 5: "Authenticate with LinkedIn"

3. Open the provided URL in your browser

4. Log in to LinkedIn and authorize the application

5. You'll be redirected back and see a success message

## Step 7: Configure Timezone (Optional)

The application defaults to UTC+7 (Asia/Bangkok) timezone. To change this:

1. Edit the `timezone` section in `config.json`:
   ```json
   "timezone": {
     "location": "America/New_York",
     "offset": "-05:00"
   }
   ```

2. Common timezone locations:
   - `Asia/Bangkok` - UTC+7 (Thailand)
   - `America/New_York` - UTC-5/-4 (Eastern Time)
   - `America/Los_Angeles` - UTC-8/-7 (Pacific Time)
   - `Europe/London` - UTC+0/+1 (GMT/BST)
   - `Asia/Tokyo` - UTC+9 (Japan)

3. The application will:
   - Display all times in your configured timezone
   - Parse schedule times as local to your timezone
   - Store times with proper timezone information

## Step 8: Start Posting!

Now you can:
- Schedule posts (option 1)
- Publish specific posts to LinkedIn (option 6)
- Auto-publish all due posts (option 7)

## Troubleshooting

### "Your LinkedIn Network Will Be Back Soon" Error
This is the most common error and usually indicates:

1. **Invalid Client ID**: 
   - Double-check your Client ID in `config.json`
   - Verify the app exists in LinkedIn Developer Portal
   - Make sure you copied the entire Client ID without extra spaces

2. **App Status Issues**:
   - Check if your LinkedIn app is active (not deleted/suspended)
   - Log into LinkedIn Developer Portal and verify app status

3. **Use the Debug Tool**:
   - Run the app and select option 8: "Debug LinkedIn authentication"
   - This will validate your configuration and show detailed auth URL info

### "Invalid redirect_uri" Error
- Make sure the redirect URL in your LinkedIn app matches exactly: `http://localhost:8080/callback`
- Check both the `config.json` file and LinkedIn app settings
- URLs are case-sensitive and must match exactly

### "Insufficient permissions" Error
- Verify that your LinkedIn app has the `w_member_social` scope enabled
- Some scopes may require LinkedIn review/approval
- Check the "Auth" tab in your LinkedIn app for enabled scopes

### "Token expired" Error
- Re-authenticate using option 5 in the app menu
- Or use the callback server at `http://localhost:8080`

### "App not approved" Error
- Some LinkedIn API features require app review
- For personal use, basic posting should work without review
- Check if your app needs verification in the LinkedIn Developer Portal

### General Debugging Steps

1. **Use the built-in debugger**:
   ```bash
   make run
   # Select option 8: Debug LinkedIn authentication
   ```

2. **Verify LinkedIn app configuration**:
   - Go to [LinkedIn Developer Portal](https://developer.linkedin.com/)
   - Check your app status and settings
   - Ensure redirect URL matches exactly

3. **Check configuration file**:
   - Verify `config.json` has correct Client ID and Secret
   - Ensure no extra spaces or characters
   - Client ID should be exactly as shown in LinkedIn portal

4. **Test with callback server**:
   ```bash
   make run-callback  # In separate terminal
   # Then open http://localhost:8080
   ```

## API Limits

- LinkedIn has rate limits on API calls
- Personal apps are limited in the number of posts per day
- For production use, consider applying for LinkedIn's Marketing API

## Security Notes

- Never commit your `config.json` or `linkedin_token.json` files to git
- These files contain sensitive credentials
- The app automatically adds them to `.gitignore`