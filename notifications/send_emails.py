from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build, Resource
from google.auth.transport.requests import Request
from googleapiclient.errors import HttpError
from google_auth_oauthlib.flow import InstalledAppFlow

from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart

import os.path
import time
import base64
from typing import Optional

# Use your Google Service Account credentials to authenticate with Google
# Define the scopes for accessing the Gmail API
scope_gmail = [
    'https://www.googleapis.com/auth/gmail.readonly',
    'https://www.googleapis.com/auth/gmail.modify',
    'https://www.googleapis.com/auth/gmail.send',
    'https://www.googleapis.com/auth/gmail.compose'
]


def get_service_gmail():
    # Set up the Gmail API credentials
    creds_path_gmail = 'client_secret.json'
    creds_gmail = None
    if os.path.exists(creds_path_gmail):
        creds_gmail = Credentials.from_authorized_user_file(creds_path_gmail, scope_gmail)
    if not creds_gmail or not creds_gmail.valid:
        if creds_gmail and creds_gmail.expired and creds_gmail.refresh_token:
            creds_gmail.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file(
                'gmail_creds.json', scope_gmail
            )
            creds = flow.run_local_server(port=0, access_type='offline')
            with open(creds_path_gmail, 'w') as token:
                token.write(creds.to_json())
            creds_gmail = Credentials.from_authorized_user_file(creds_path_gmail, scope_gmail)

    if not creds_gmail or not creds_gmail.valid:
        raise Exception('Invalid creds ' + creds_path_gmail)

    service_gmail = build('gmail', 'v1', credentials=creds_gmail)
    return service_gmail


def fetch_emails(service_gmail: Resource, query: Optional[str] = ''):
    #  query = f"subject:\"{SUBJECT}\" to:{email}"
    if not query:
        response = service_gmail.users().messages().list(userId="me").execute()
    else:
        response = service_gmail.users().messages().list(userId="me", q=query).execute()

    emails = response.get('messages', [])
    if not emails:
        emails = []

    for email in emails:
        msg = service_gmail.users().messages().get(userId='me', id=email['id']).execute()
        date = msg['internalDate']
        headers = msg['payload']['headers']
        sender = [header['value'] for header in headers if header['name'] == 'From'][0]
        subject = [header['value'] for header in headers if header['name'] == 'Subject'][0]
        to = [
            header['value'] for header in headers if header['name'] == 'To'
        ]
        cc = [
            header['value'] for header in headers if header['name'] == 'Cc'
        ]
        bcc = [
            header['value'] for header in headers if header['name'] == 'Bcc'
        ]
        print((date, sender, to, cc, bcc, subject))

    return response


def send_email(service_gmail: Resource, to):
    # Create a MIME Multipart message
    message = MIMEMultipart()

    # Add the subject
    message['subject'] = f'Dernier rappel: Tournoi de Lognes 28-29/10/2023'

    # Add the recipient(s)
    message['from'] = f'eplognes <tournoiseplognes@gmail.com>'
    #  message['from'] = f'Jean-Christophe KOURAJIAN <{OVH_EMAIL}>'
    message['to'] = to

    with open('email_template.html', 'r') as f:
        message_html = f.read()

    # Create a MIMEText object for the HTML message
    html_message = MIMEText(message_html, 'html')
    message.attach(html_message)

    # Base64 encode the message
    raw_message = base64.urlsafe_b64encode(message.as_bytes()).decode("utf-8")

    # Send the email
    try:
        message = service_gmail.users().messages().send(
            userId="me",
            body={'raw': raw_message}
        ).execute()
        print(f"Message sent. Message ID: {message['id']}")
    except HttpError as error:
        print(f"An error occurred: {error}")


service_gmail = get_service_gmail()
"""
"""
send_email(service_gmail, 'aurelienduboc96@gmail.com')
exit(0)

with open('to.txt', 'r') as f:
    EMAILS = [
        item.strip()
        for item in f.read().split('\n')
        if item
    ]

print('----')
for key, email in enumerate(EMAILS):
    print(1+key, email)
    send_email(service_gmail, email)
    print('----')
    time.sleep(1)
