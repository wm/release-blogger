# Release Blogger

## Usage

Run the web server that listens for events from Github web hooks as follows:

    release-blogger \
      --port 8080 \
      --secret my-github-hmac-secret \
      --space BLOG \
      --url https://mycompany.atlassian.net \
      --username my-confluence-username \
      --password my-confluence-password


In github set up your webhook to point to url of the release-blogger server
with a patt of `event`. For Example:

    http://release-blogger.ngrok.io/event

When ever a Github Release Event is pushed to the release-blogger it will
create a new blog entry in the provided space of confluence.
