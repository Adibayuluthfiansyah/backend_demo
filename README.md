# API Users

POST /api/users/admin
Input:
{
  "name": "string",
  "username": "string",
  "password": "string"
}
Response (201 Created):
{
  "message": "Admin berhasil dibuat",
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string",
    "role": "admin"
  }
}

POST /api/users/staff
Input:
{
  "name": "string",
  "username": "string",
  "password": "string"
}
Response (201 Created):
{
  "message": "Staff berhasil dibuat",
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string",
    "role": "staff"
  }
}

GET /api/users
Input: -
Response (200 OK):
[
  {
    "id": "uuid",
    "name": "string",
    "username": "string",
    "role": "admin/staff"
  }
]

GET /api/users/:id
Input: -
Response (200 OK):
{
  "id": "uuid",
  "name": "string",
  "username": "string",
  "role": "admin/staff"
}
Response (404 Not Found):
{
  "error": "User tidak ditemukan"
}

PUT /api/users/:id
Input:
{
  "name": "string (opsional)",
  "username": "string (opsional)",
  "password": "string (opsional)"
}
Response (200 OK):
{
  "message": "User berhasil diperbarui",
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string",
    "role": "admin/staff"
  }
}
Response (404 Not Found):
{
  "error": "User tidak ditemukan"
}

DELETE /api/users/:id
Input: -
Response (200 OK):
{
  "message": "User berhasil dihapus"
}
Response (404 Not Found):
{
  "error": "User tidak ditemukan"
}



# API Login

POST /api/login
Input:
{
  "username": "string",
  "password": "string"
}
Response (200 OK):
{
  "message": "Login berhasil",
  "token_id": "uuid"
}
Response (400 Bad Request):
{
  "message": "Input tidak valid"
}
Response (401 Unauthorized):
{
  "message": "Username atau password salah"
}
Response (500 Internal Server Error):
{
  "message": "Gagal membuat token"
}




# API Logout

POST /api/logout
Input:
{
  "token_id": "uuid"
}
Response (200 OK):
{
  "message": "Logout berhasil"
}
Response (400 Bad Request):
{
  "message": "Input tidak valid"
}
Response (404 Not Found):
{
  "message": "Token tidak ditemukan"
}
Response (500 Internal Server Error):
{
  "message": "Gagal logout"
}




# API Documents

POST /api/documents
Input (multipart/form-data):
- sender: string
- subject: string
- letter_type: string
- file: PDF atau gambar (jpg, jpeg, png, gif, webp)
Response (201 Created):
{
  "document": {
    "id": "uuid",
    "sender": "string",
    "subject": "string",
    "letter_type": "string",
    "file_name": "url_file",
    "public_id": "cloudinary_id",
    "resource_type": "pdf/image",
    "user_id": "uuid"
  }
}
Response (400 Bad Request):
{
  "error": "File tidak ditemukan / format file tidak didukung"
}
Response (401 Unauthorized):
{
  "error": "User tidak terautentikasi"
}
Response (500 Internal Server Error):
{
  "error": "Tidak dapat membuka file / Upload gagal / DB error"
}

GET /api/documents
Input: -
Response (200 OK):
{
  "documents": [
    {
      "id": "uuid",
      "sender": "string",
      "subject": "string",
      "letter_type": "string",
      "file_name": "url_file",
      "public_id": "cloudinary_id",
      "resource_type": "pdf/image",
      "user_id": "uuid",
      "user": { "id": "uuid", "name": "string", "username": "string", "role": "admin" }
    }
  ]
}

GET /api/documents/:id
Input: -
Response (200 OK):
{
  "document": {
    "id": "uuid",
    "sender": "string",
    "subject": "string",
    "letter_type": "string",
    "file_name": "url_file",
    "public_id": "cloudinary_id",
    "resource_type": "pdf/image",
    "user_id": "uuid",
    "user": { "id": "uuid", "name": "string", "username": "string", "role": "admin" }
  }
}
Response (404 Not Found):
{
  "error": "Dokumen tidak ditemukan"
}

PUT /api/documents/:id
Input (multipart/form-data, opsional file):
- sender: string (opsional)
- subject: string (opsional)
- letter_type: string (opsional)
- file: PDF atau gambar (opsional)
Response (200 OK):
{
  "message": "Dokumen berhasil diperbarui",
  "document": {
    "id": "uuid",
    "sender": "string",
    "subject": "string",
    "letter_type": "string",
    "file_name": "url_file",
    "public_id": "cloudinary_id",
    "resource_type": "pdf/image",
    "user_id": "uuid"
  }
}
Response (400 Bad Request):
{
  "error": "Format file tidak didukung"
}
Response (404 Not Found):
{
  "error": "Dokumen tidak ditemukan"
}
Response (500 Internal Server Error):
{
  "error": "Gagal menyimpan perubahan / Upload gagal"
}

DELETE /api/documents/:id
Input: -
Response (200 OK):
{
  "message": "Dokumen berhasil dihapus"
}
Response (404 Not Found):
{
  "error": "Dokumen tidak ditemukan"
}



# API Document Staff

POST /api/document_staff
Input (multipart/form-data):
- subject: string
- file: PDF atau gambar (jpg, jpeg, png, gif, webp)
Response (201 Created):
{
  "document_staff": {
    "id": "uuid",
    "user_id": "uuid",
    "subject": "string",
    "file_name": "url_file",
    "public_id": "cloudinary_id",
    "resource_type": "pdf/image"
  }
}
Response (400 Bad Request):
{
  "error": "File tidak ditemukan / format file tidak didukung"
}
Response (401 Unauthorized):
{
  "error": "User tidak terautentikasi"
}
Response (500 Internal Server Error):
{
  "error": "Tidak dapat membuka file / Upload gagal / DB error"
}

GET /api/document_staff
Input: -
Response (200 OK):
{
  "document_staffs": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "subject": "string",
      "file_name": "url_file",
      "public_id": "cloudinary_id",
      "resource_type": "pdf/image",
      "user": { "id": "uuid", "name": "string", "username": "string", "role": "admin/staff" }
    }
  ]
}

GET /api/document_staff/:id
Input: -
Response (200 OK):
{
  "document_staff": {
    "id": "uuid",
    "user_id": "uuid",
    "subject": "string",
    "file_name": "url_file",
    "public_id": "cloudinary_id",
    "resource_type": "pdf/image",
    "user": { "id": "uuid", "name": "string", "username": "string", "role": "admin/staff" }
  }
}
Response (404 Not Found):
{
  "error": "Dokumen tidak ditemukan"
}

PUT /api/document_staff/:id
Input (multipart/form-data, opsional file):
- subject: string (opsional)
- file: PDF atau gambar (opsional)
Response (200 OK):
{
  "message": "Dokumen berhasil diperbarui",
  "document_staff": {
    "id": "uuid",
    "user_id": "uuid",
    "subject": "string",
    "file_name": "url_file",
    "public_id": "cloudinary_id",
    "resource_type": "pdf/image"
  }
}
Response (400 Bad Request):
{
  "error": "Format file tidak didukung"
}
Response (404 Not Found):
{
  "error": "Dokumen tidak ditemukan"
}
Response (500 Internal Server Error):
{
  "error": "Gagal menyimpan perubahan / Upload gagal"
}

DELETE /api/document_staff/:id
Input: -
Response (200 OK):
{
  "message": "Dokumen berhasil dihapus"
}
Response (404 Not Found):
{
  "error": "Dokumen tidak ditemukan"
}




# API Superior Orders

POST /api/superior_orders
Input:
{
  "document_id": "uuid",
  "user_ids": ["uuid1", "uuid2", "..."]
}
Response (201 Created):
{
  "message": "SuperiorOrder created",
  "data": [
    {
      "document_id": "uuid",
      "user_id": "uuid1",
      "id": "uuid"
    },
    {
      "document_id": "uuid",
      "user_id": "uuid2",
      "id": "uuid"
    }
  ]
}
Response (400 Bad Request):
{
  "error": "Invalid input: ..."
}
Response (500 Internal Server Error):
{
  "error": "Failed to create record: ..."
}

GET /api/superior_orders
Input: -
Response (200 OK):
{
  "data": [
    {
      "document_id": "uuid",
      "sender": "string",
      "subject": "string",
      "users": [
        { "name": "string" },
        { "name": "string" }
      ]
    }
  ]
}
Response (500 Internal Server Error):
{
  "error": "Failed to fetch records: ..."
}

GET /api/superior_orders/:document_id
Input: -
Response (200 OK):
{
  "document_id": "uuid",
  "user_ids": [
    { "document_id": "uuid", "user_id": "uuid1", "id": "uuid" },
    { "document_id": "uuid", "user_id": "uuid2", "id": "uuid" }
  ]
}
Response (404 Not Found):
{
  "error": "No records found for this document"
}
Response (500 Internal Server Error):
{
  "error": "Failed to fetch records: ..."
}

PUT /api/superior_orders/:document_id
Input:
{
  "user_ids": ["uuid1", "uuid2", "..."]
}
Response (200 OK):
{
  "message": "SuperiorOrder updated",
  "data": [
    { "document_id": "uuid", "user_id": "uuid1", "id": "uuid" },
    { "document_id": "uuid", "user_id": "uuid2", "id": "uuid" }
  ]
}
Response (400 Bad Request):
{
  "error": "Invalid input: ..."
}
Response (500 Internal Server Error):
{
  "error": "Failed to delete old records / Failed to create record: ..."
}

DELETE /api/superior_orders/:document_id
Input: -
Response (200 OK):
{
  "message": "All SuperiorOrders for document deleted",
  "document_id": "uuid"
}
Response (500 Internal Server Error):
{
  "error": "Failed to delete records: ..."
}
