# Email Setup Guide

This document explains how to configure email functionality in Recontext platform.

## Overview

The platform sends welcome emails to newly registered users in their preferred language. The email system supports:
- Multi-language email templates (English and Russian)
- User language preferences
- Asynchronous email sending
- SMTP configuration via environment variables

## SMTP Configuration

### Environment Variables

Set the following environment variables in your `.env` file or docker-compose configuration:

```bash
# SMTP Server Configuration
SMTP_HOST=smtp.gmail.com          # SMTP server hostname
SMTP_PORT=587                      # SMTP port (587 for TLS, 465 for SSL)
SMTP_USER=your-email@example.com  # SMTP username (usually your email)
SMTP_PASSWORD=your-password        # SMTP password or app-specific password
SMTP_FROM_ADDRESS=noreply@recontext.online  # From email address
SMTP_FROM_NAME=Recontext           # From display name

# Optional: Login URL for email templates
LOGIN_URL=https://your-domain.com  # URL shown in welcome emails
```

### Supported Email Providers

#### Gmail
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-specific-password
```

**Note**: For Gmail, you need to:
1. Enable 2-factor authentication
2. Generate an app-specific password at: https://myaccount.google.com/apppasswords

#### Mail.ru
```bash
SMTP_HOST=smtp.mail.ru
SMTP_PORT=465
SMTP_USER=your-email@mail.ru
SMTP_PASSWORD=your-password
```

#### Yandex
```bash
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=465
SMTP_USER=your-email@yandex.ru
SMTP_PASSWORD=your-app-password
```

#### Timeweb (Current Production)
```bash
SMTP_HOST=smtp.timeweb.ru
SMTP_PORT=465
SMTP_USER=support@recontext.online
SMTP_PASSWORD=your-password
```

## Docker Deployment

The SMTP configuration is already included in `docker-compose.yml` and `docker-compose.prod.yml`.

### Production Deployment

1. Create a `.env` file in the project root:
```bash
cp .env.example .env
```

2. Edit `.env` and set your SMTP credentials:
```bash
SMTP_HOST=smtp.your-provider.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASSWORD=your-password
SMTP_FROM_ADDRESS=noreply@recontext.online
SMTP_FROM_NAME=Recontext
LOGIN_URL=https://your-domain.com
```

3. Start the services:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Email Templates

Email templates are located in `pkg/email/templates/`:
- `welcome_en.html` - English welcome email
- `welcome_ru.html` - Russian welcome email

The appropriate template is automatically selected based on the user's language preference.

### Template Variables

Welcome email templates support the following variables:
- `{{.Username}}` - User's username
- `{{.Email}}` - User's email address
- `{{.Language}}` - User's preferred language
- `{{.LoginURL}}` - URL to login page

## User Language Preferences

### Backend

Users have a `language` field in the database:
- Default value: `"en"`
- Supported values: `"en"`, `"ru"`
- Can be updated via API: `PUT /api/v1/users/{id}`

### Frontend

When a user logs in:
1. The backend returns the user's language preference
2. The frontend automatically switches to that language
3. The preference is saved in localStorage

Users can change their language in the User Settings page:
- Path: `/settings` (if routed)
- Component: `UserSettings.tsx`

## Testing Email Functionality

### 1. Test Email Sending

Create a test user via API:
```bash
curl -X POST http://localhost:20080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "language": "ru"
  }'
```

A welcome email should be sent to `test@example.com` in Russian.

### 2. Check Logs

Monitor logs for email sending:
```bash
docker logs recontext-managing-portal | grep -i email
```

Successful email: `Welcome email sent to test@example.com`
Failed email: `Failed to send welcome email to test@example.com: <error>`

## Troubleshooting

### Email Not Sending

1. **Check SMTP credentials**: Verify environment variables are set correctly
2. **Check firewall**: Ensure outbound connections to SMTP port are allowed
3. **Check logs**: Look for error messages in container logs
4. **Test SMTP connection**:
```bash
telnet smtp.gmail.com 587
```

### Common Errors

**Authentication failed**:
- Verify username and password
- For Gmail, ensure app-specific password is used
- Check if 2FA is required

**Connection timeout**:
- Verify SMTP_HOST and SMTP_PORT
- Check firewall rules
- Verify DNS resolution

**SSL/TLS errors**:
- Use port 587 for TLS (recommended)
- Use port 465 for SSL
- Ensure SMTP_PORT matches your provider's requirements

## Security Best Practices

1. **Never commit credentials**: Use environment variables or secrets management
2. **Use app-specific passwords**: For Gmail and similar providers
3. **Encrypt in transit**: Always use TLS/SSL
4. **Rotate passwords**: Regularly update SMTP passwords
5. **Monitor usage**: Watch for unusual email sending patterns

## Email Features

- ✅ Welcome emails on user registration
- ✅ Multi-language support (EN, RU)
- ✅ HTML email templates
- ✅ Asynchronous sending (non-blocking)
- ✅ User language preference
- ⏳ Password reset emails (planned)
- ⏳ Meeting notifications (planned)
- ⏳ Recording ready notifications (planned)

## Migration

A database migration has been added to support user language preferences:
```sql
-- migration: 000012_add_language_to_users
ALTER TABLE users ADD COLUMN language VARCHAR(10) NOT NULL DEFAULT 'en';
```

The migration runs automatically when the managing-portal starts.
