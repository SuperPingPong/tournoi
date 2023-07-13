from flask import Flask, request, jsonify, send_file
from flask_cors import CORS

from datetime import datetime
import json
import jwt
import os
import psycopg2
import psycopg2.extras

import gspread
from oauth2client.service_account import ServiceAccountCredentials
import utils

app = Flask(__name__)
CORS(app)
app.config['CORS_HEADERS'] = 'Content-Type'

# Debug
debug = os.environ.get('DEBUG', False)
app.debug = debug
admin_email = os.environ.get('ADMIN_EMAIL', 'admin@example.com')

# Database connection details
db_host = "db"
db_name = "database"
db_user = os.environ.get('POSTGRES_USERNAME', 'postgres')
db_password = os.environ.get('POSTGRES_PASSWORD', 'postgres')

# Establish a connection to the database
dsn = f"host={db_host} user={db_user} password={db_password} dbname={db_name} port=5432 sslmode=disable"
conn = psycopg2.connect(dsn)

# Set the secret key used for signing the JWT
jwt_secret_key = os.environ.get('JWT_SECRET_KEY', 'secret')

# Create a cursor object to execute SQL queries
db = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

# Perform the join query
query = "SELECT day, name, max_entries FROM bands"
db.execute(query)
BANDS = list(map(dict, db.fetchall()))
BANDS = {
    band['name']: {
        'index': key,
        'day': band['day'],
        'max_entries': band['max_entries'],
    } for key, band in enumerate(BANDS)
}

# Use your Google Service Account credentials to authenticate with Google
scope = ['https://spreadsheets.google.com/feeds', 'https://www.googleapis.com/auth/drive']
client_secret_json = os.environ.get('CLIENT_SECRET_JSON')
client_secret_dict = json.loads(client_secret_json)
credentials = ServiceAccountCredentials.from_json_keyfile_dict(client_secret_dict, scope)
gc = gspread.authorize(credentials)


def get_email_from_db(uid):
    query = f"SELECT email FROM users WHERE id = '{uid}'"
    db.execute(query)
    result = db.fetchone()
    if result:
        return result.get('email')
    return None


def extract_email_from_jwt(jwt_cookie):
    if not jwt_cookie:
        return jsonify({'error': 'Invalid JWT'}), 401

    # Extract the user_id (uid) from the JWT
    decoded_token = jwt.decode(jwt_cookie, jwt_secret_key, algorithms=["HS256"], verify=False)
    uid = decoded_token.get('uid')

    if not uid:
        return jsonify({'error': 'Invalid JWT'}), 401

    # Retrieve the email from the database based on the uid
    email = get_email_from_db(uid)
    if not email:
        return jsonify({'error': 'Invalid JWT'}), 401

    if email != os.environ.get('ADMIN_EMAIL'):
        return jsonify({'error': 'Access forbidden'}), 403

    return email, 200


@app.route("/api/export", methods=['GET', 'OPTIONS'])
def export():
    # Get the JWT from the Cookie header
    jwt_cookie = request.cookies.get('jwt')
    response, status_code = extract_email_from_jwt(jwt_cookie)
    if status_code != 200:
        return response, status_code

    client_secret_json = os.environ.get('CLIENT_SECRET_JSON')
    client_secret_dict = json.loads(client_secret_json)
    credentials = ServiceAccountCredentials.from_json_keyfile_dict(client_secret_dict, scope)
    gc = gspread.authorize(credentials)

    # Open the target Google Sheet by its url
    spreadsheet_id = '1hA2P_yKZKJ7JZ3hgZmVu-Um_pPBgjVIHzaq2yeqrwrA'
    spreadsheet_url = f"https://docs.google.com/spreadsheets/d/{spreadsheet_id}"
    workbook = gc.open_by_url(spreadsheet_url)
    worksheet = workbook.worksheet("Base inscrits")

    query = """
            SELECT users.email, 
            members.last_name,
            members.first_name,
            members.permit_id,
            members.club_name,
            members.points::int,
            members.category,
            bands.name as band_name,
            bands.max_entries AS band_max_entries,
            users.email
            FROM entries
            JOIN members ON entries.member_id = members.id
            JOIN users ON members.user_id = users.id
            JOIN bands ON entries.band_id = bands.id
            WHERE entries.confirmed is TRUE
            AND entries.deleted_at is NULL
            ORDER BY entries.created_at ASC
        """
    try:
        db.execute(query)
        entries = db.fetchall()
    except Exception as _:
        error_message = "An error occurred during the database query."
        return jsonify({"error": error_message}), 500

    utils.clean_worksheet(worksheet)
    utils.fill_worksheet(worksheet, BANDS, entries)

    # Download the workbook as XLSX
    workbook_as_xlsx = workbook.export(format='application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')

    # Save the exported workbook to a file
    timestamp = datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
    exported_file_path = f'{timestamp}-tournoi-de-lognes.xlsx'

    with open(exported_file_path, 'wb') as f:
        f.write(workbook_as_xlsx)

    # Send the file to the user for download
    return send_file(exported_file_path, as_attachment=True)


if __name__ == "__main__":
    app.run(host="0.0.0.0", debug=debug)
