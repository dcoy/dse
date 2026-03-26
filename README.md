# WorkOS 🤝 Okta - SSO + Directory Sync

An example Go application demonstrating how to:
  - Authenticate users using WorkOS SSO (via Okta)
  - Sync and display users using WorkOS Directory Sync (SCIM)

## Screen Recording 

Short and sweet recording demonstrating SSO auth + directory users listing: [DSE App in action](https://youtu.be/9czx7ldxLDE)

## Prerequisites

- Go 1.16+
- [WorkOS Dashboard account](https://dashboard.workos.com/signup)
- [Okta Developer account](https://developer.okta.com/signup/)
  - Select "Integrator Free Plan"
  - Use a work/school email

In order to enable the project to run successfully, both the WorkOS and Okta accounts should be configured.

## SSO Setup with WorkOS

1. Create a WorkOS organization
2. under single sign-on, click the "Configure manually" button
3. Choose Okta SAML from the dropdown
4. Click the "Add connection" button
5. Set up the redirect URI under [Redirects](https://dashboard.workos.com/redirects) to `http://localhost:8000/callback`

## Okta SAML App Setup

Okta provides a guide to[Create SAML app integrations](https://help.okta.com/oie/en-us/content/topics/apps/apps_app_integration_wizard_saml.htm) to setup a SAML app in your Okta account. Here's a condensed version:

In a new browser tab:

1. Sign-in to your Okta organization
2. On the left, expand Applications and click "Applications"
3. Click the "Create App Integration" button and choose SAML 2.0
4. Give your SAML app integration a name
5. Update the following with the details from WorkOS:

   - Single sign-on URL
      - In WorkOS, this is the ACS URL for your standalone SSO connection
   - Audience URI (SP Entity ID)
      - In WorkOS, this is the SP Entity ID for your standalone SSO connection
   - NameID format: EmailAddress
   - Application username: Email

6. Click "Next"
7. Select "This is an internal app that we have created"
8. Copy the "Metadata URL"
9. Move back to WorkOS
10. Under "Identity Provider Configuration", click on the "Edit Configuration" button
11. Paste the copied Metadata URL from your Okta SAML app and save

### Attribute mapping

In order for WorkOS to properly ingest Okta SAML attributes, you'll need to map the attributes in Okta, so they're sent in the expected format to WorkOS. To do this:

1. Access your newly created Okta SAML App under "Applications"
2. Access the "Sign On" tab
3. Scroll to Attribute statements and expand "Show legacy configuration"
4. Click "Edit" and enter the attribute statements exactly as follows, then save:

| Name | Name format | Value |
| :--- | :---: | :--- |
| email | Basic | user.email |
| firstName | Basic | user.firstName |
| idpId | Basic | user.NameID |
| lastName | Basic | user.lastName |

## WorkOS Directory Sync

There's a bit of jumping between Okta and WorkOS, so open up two tabs to access your WorkOS SSO connection and Okta SAML app.

In **WorkOS**:

1. Scroll to User Provisioning and click "Configure manually"
2. Choose Okta and give your directory a name

In **Okta**:

1. Open up your Okta SAML App, and in the General tab, edit the App settings
2. Choose "SCIM" in Provisioning and save
3. In the new Provisioning tab, edit the "SCIM Connection"
4. With the values from your WorkOS directory connection, enter the following and save:
   - SCIM connector base URL:  the WorkOS directory _Endpoint_
   - Unique identifier: email
   - Supported provision actions:
      - Import New Users and Profile Updates
      - Push New Users
      - Push Profile Updates
   - Authentication mode: HTTP Header
   - Authorization Bearer token: WorkOS _Bearer Token_
5. Choose the following for _Provisioning to App_ and save:
   - Create Users
   - Update User Attributes
   - Deactivate Users

Now that you have your SSO connection and directory sync configured, in Okta create some users in your Directory > People, and set them to active. From there, you can add them to your SAML application under the _Assignments_ tab in your SAML app. Be sure you add yourself to the app, otherwise, you won't be able to sign-in! 

Finally, you can validate that directory sync is working by heading to your Directory connection > Users. 


## Go Project Setup

1. Clone this repository using your preferred secure method (HTTPS, SSH, or Github CLI)

```bash
# HTTPS
git clone https://github.com/dcoy/dse.git

# SSH
git clone git@github.com:dcoy/dse.git

# GitHub CLI
gh repo clone dcoy/dse
```

2. Navigate to the cloned repository

```bash
cd dse
```

3. Install dependencies

```shell
go mod tidy
```

4. Create a new file called ".env" in the project root and add the following variables:

- WorkOS API key and Client ID: [API keys](https://dashboard.workos.com/api-keys)
- WorkOS Connection ID: Located in your [WorkOS Dashboard](https://dashboard.workos.com/dashboard) > Organization > View SSO connection button > Connection ID (located in the banner at the top)
- WorkOS Directory ID: Located in your [WorkOS Dashboard](https://dashboard.workos.com/dashboard) > Organization > View directory button > Directory ID (located in the banner at the top)

```shell
WORKOS_API_KEY=your_api_key
WORKOS_CLIENT_ID=your_client_id
WORKOS_REDIRECT_URI=http://localhost:8000/callback
WORKOS_CONNECTION=your_connection_id
WORKOS_DIRECTORY_ID=your_directory_id
```

5. The final setup step is to start the server

   ```bash
   go run .
   ```

   Once running, navigate to [http://localhost:8000] to test out the SSO workflow. You will see your first and last name on the page, and you click on the _Directory Users_ button to view a list of the directory users.

## Troubleshooting

### 404 after clicking Login

- Ensure `WORKOS_CONNECTION` is set correctly
- Confirm SSO connection is active in WorkOS

### 422 Validation Error (Directory Sync)

- Verify `WORKOS_DIRECTORY_ID`
- Ensure users exist and are assigned in Okta

### Missing user attributes

- Confirm [attribute mappings](#attribute-mapping) in Okta:
   - email
   - firstName
   - idpId
   - lastName

## Help

If the [Troubleshooting](#troubleshooting) section doesn't cover anything you've run into, feel free to [create an issue](https://github.com/dcoy/dse/issues/new) with the `help wanted` label in this repository.