from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build, Resource
from google.auth.transport.requests import Request
from googleapiclient.errors import HttpError
from google_auth_oauthlib.flow import InstalledAppFlow

from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders

import os.path
from pathlib import Path
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


def send_email(service_gmail: Resource, to, template: str, attachment_path: Optional[str] = None):
    # Create a MIME Multipart message
    message = MIMEMultipart()

    # Add the subject
    #  message['subject'] = f'Ouverture des inscriptions: Tournoi de Lognes 08-09/06/2024'
    message['subject'] = f'Information importante: Tournoi de Lognes 26-27/10/2024'

    # Add the recipient(s)
    message['from'] = f'eplognes <tournoiseplognes@gmail.com>'
    message['to'] = to

    email_template_path = Path(f'templates/{template}/email_template.html')
    if email_template_path.exists():
        message_html = email_template_path.read_text()
    else:
        raise Exception(f"File does not exist: {email_template_path}")

    # Create a MIMEText object for the HTML message
    html_message = MIMEText(message_html, 'html')
    message.attach(html_message)

    if 'EXTERNAL_URL' in html_message:
        raise Exception('Please replace EXTERNAL_URL from template')

    # Attach the file if provided
    if attachment_path:
        # Open the file in binary mode
        with open(attachment_path, 'rb') as attachment:
            # Create a MIMEBase object
            #  mime_base = MIMEBase('application', 'octet-stream')
            mime_base = MIMEBase('application', 'vnd.openxmlformats-officedocument.presentationml.presentation')
            mime_base.set_payload(attachment.read())

        # Encode the payload using base64
        encoders.encode_base64(mime_base)

        # Add header to the MIMEBase object
        mime_base.add_header(
            'Content-Disposition',
            f'attachment; filename={os.path.basename(attachment_path)}'
        )

        # Attach the MIMEBase object to the MIMEMultipart message
        message.attach(mime_base)

    # Base64 encode the message
    raw_message = base64.urlsafe_b64encode(message.as_bytes()).decode("utf-8")

    # Send the email
    try:
        message = service_gmail.users().messages().send(
            userId="me",
            body={'raw': raw_message}
        ).execute()
        print(f"Message sent. Message ID: {message['id']} - {to}")
    except HttpError as error:
        print(f"An error occurred: {error}")


service_gmail = get_service_gmail()

"""
c=0; for i in backup/*; do c=$((c+1)); ./fetch_emails.sh $i; mv to.txt emails/$c.txt; done
rm all_emails.txt 2>/dev/null
for f in emails/*; do cat $f>>all_emails.txt; done
sort -u all_emails.txt>temp
mv temp all_emails.txt
wc -l all_emails.txt
"""
with open('all_emails.txt', 'r') as f:
    PREVIOUSLY_REGISTERED_EMAILS = [
        item.strip()
        for item in f.read().split('\n')
        if item
    ]

with open('to.txt', 'r') as f:
    ALREADY_REGISTERED_EMAILS = [
        item.strip()
        for item in f.read().split('\n')
        if item
    ]
EMAILS = [e for e in PREVIOUSLY_REGISTERED_EMAILS if e not in ALREADY_REGISTERED_EMAILS]
TEMPLATE='tournament_open'
send_email(service_gmail, 'aurelienduboc96@gmail.com', template=TEMPLATE)

exit(0)
#  print(len(EMAILS))
with open('to.txt', 'r') as f:
    EMAILS = [
        item.strip()
        for item in f.read().split('\n')
        if item
    ]
print(len(EMAILS))
print('----')
#  EMAILS = ['aurelienduboc96@gmail.com']
ATTACHMENT_PATH='./attachments/Info-joueurs-tarif-juin.pptx'

for key, email in enumerate(EMAILS):
    print(1+key, email)
    send_email(service_gmail, email, template=TEMPLATE, attachment_path=ATTACHMENT_PATH)
    print('----')
    time.sleep(1)
