from flask import Flask, request, jsonify
from flask_cors import CORS

import jwt
import psycopg2
import psycopg2.extras

import os

app = Flask(__name__)
CORS(app)
app.config['CORS_HEADERS'] = 'Content-Type'

# Debug
debug = os.environ.get('DEBUG', False)
app.debug = debug

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

    return email, 200


@app.route("/api/export", methods=['GET', 'OPTIONS'])
def export():
    # Get the JWT from the Cookie header
    jwt_cookie = request.cookies.get('jwt')
    response, status_code = extract_email_from_jwt(jwt_cookie)
    if status_code != 200:
        return response, status_code

    email = response
    print(email)
    print(email)

    # Perform the join query
    query = """
            SELECT entries.*, members.*, users.*
            FROM entries
            JOIN members ON entries.member_id = members.id
            JOIN users ON members.user_id = users.id
        """
    db.execute(query)
    results = db.fetchall()
    """
    for row in results:
        print(row)
        pass
    """
    return jsonify(results), 200
    try:
        # Perform the join query
        query = """
            SELECT entries.*, members.*, users.*
            FROM entries
            JOIN members ON entries.member_id = members.id
            JOIN users ON members.user_id = users.id
        """
        db.execute(query)

        # Fetch all the results
        results = db.fetchall()

        # Format the results
        formatted_results = []
        for row in results:
            formatted_results.append({
                'entry_id': row[0],
                'entry_data': row[1],
                'member_id': row[2],
                'member_data': row[3],
                'user_id': row[4],
                'user_data': row[5]
            })

        # Return the formatted results as JSON
        return jsonify(formatted_results[:10]), 200

    except psycopg2.Error as e:
        # Handle any database errors
        return jsonify({'error': str(e)}), 500

    finally:
        # Close the cursor and the database connection
        db.close()
        conn.close()


if __name__ == "__main__":
    app.run(host="0.0.0.0", debug=debug)
