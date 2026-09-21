-- name: CreateConversation :one
INSERT INTO conversations (type, title, created_by, wa_contact_phone, wa_contact_name, resident_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetConversationByID :one
SELECT *
FROM conversations
WHERE id = $1;

-- name: GetWhatsAppConversationByPhone :one
SELECT *
FROM conversations
WHERE type = 'WHATSAPP' AND wa_contact_phone = $1;

-- name: FindDirectConversationBetween :one
SELECT c.*
FROM conversations c
INNER JOIN conversation_participants p1
    ON p1.conversation_id = c.id AND p1.user_id = $1
INNER JOIN conversation_participants p2
    ON p2.conversation_id = c.id AND p2.user_id = $2
WHERE c.type = 'DIRECT'
LIMIT 1;

-- name: AddConversationParticipant :exec
INSERT INTO conversation_participants (conversation_id, user_id, last_read_at)
VALUES ($1, $2, $3);

-- name: GetConversationParticipant :one
SELECT *
FROM conversation_participants
WHERE conversation_id = $1 AND user_id = $2;

-- name: ListParticipantsByConversationID :many
SELECT
    cp.conversation_id,
    u.id AS user_id,
    u.name,
    u.email
FROM conversation_participants cp
INNER JOIN users u ON u.id = cp.user_id
WHERE cp.conversation_id = $1
ORDER BY u.name ASC;

-- name: ListParticipantsByConversationIDs :many
SELECT
    cp.conversation_id,
    u.id AS user_id,
    u.name,
    u.email
FROM conversation_participants cp
INNER JOIN users u ON u.id = cp.user_id
WHERE cp.conversation_id = ANY(sqlc.arg(conversation_ids)::uuid[])
ORDER BY cp.conversation_id, u.name ASC;

-- name: ListMyConversations :many
SELECT
    c.id,
    c.type,
    c.title,
    c.created_by,
    c.wa_contact_phone,
    c.wa_contact_name,
    c.resident_id,
    r.name AS resident_name,
    c.last_message_at,
    c.last_message_preview,
    c.created_at,
    c.updated_at,
    me.last_read_at AS my_last_read_at,
    (
        SELECT COUNT(*)::bigint
        FROM messages m
        WHERE m.conversation_id = c.id
          AND (
              (c.type = 'WHATSAPP' AND m.sender_kind = 'CONTACT')
              OR (
                  c.type <> 'WHATSAPP'
                  AND m.sender_kind = 'USER'
                  AND m.sender_id IS DISTINCT FROM sqlc.arg(user_id)
              )
          )
          AND (me.last_read_at IS NULL OR m.created_at > me.last_read_at)
    ) AS unread_count
FROM conversations c
LEFT JOIN conversation_participants me
    ON me.conversation_id = c.id AND me.user_id = sqlc.arg(user_id)
LEFT JOIN residents r
    ON r.id = c.resident_id
WHERE
    (
        c.type = 'WHATSAPP'
        OR me.user_id IS NOT NULL
    )
    AND (
        sqlc.arg(search)::text = ''
        OR COALESCE(c.title, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR COALESCE(c.wa_contact_name, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR COALESCE(c.wa_contact_phone, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR COALESCE(r.name, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR c.last_message_preview ILIKE '%' || sqlc.arg(search) || '%'
        OR EXISTS (
            SELECT 1
            FROM conversation_participants cp
            INNER JOIN users u ON u.id = cp.user_id
            WHERE cp.conversation_id = c.id
              AND cp.user_id <> sqlc.arg(user_id)
              AND (
                  u.name ILIKE '%' || sqlc.arg(search) || '%'
                  OR u.email ILIKE '%' || sqlc.arg(search) || '%'
              )
        )
    )
ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountMyConversations :one
SELECT COUNT(*)::bigint
FROM conversations c
LEFT JOIN conversation_participants me
    ON me.conversation_id = c.id AND me.user_id = sqlc.arg(user_id)
LEFT JOIN residents r
    ON r.id = c.resident_id
WHERE
    (
        c.type = 'WHATSAPP'
        OR me.user_id IS NOT NULL
    )
    AND (
        sqlc.arg(search)::text = ''
        OR COALESCE(c.title, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR COALESCE(c.wa_contact_name, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR COALESCE(c.wa_contact_phone, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR COALESCE(r.name, '') ILIKE '%' || sqlc.arg(search) || '%'
        OR c.last_message_preview ILIKE '%' || sqlc.arg(search) || '%'
        OR EXISTS (
            SELECT 1
            FROM conversation_participants cp
            INNER JOIN users u ON u.id = cp.user_id
            WHERE cp.conversation_id = c.id
              AND cp.user_id <> sqlc.arg(user_id)
              AND (
                  u.name ILIKE '%' || sqlc.arg(search) || '%'
                  OR u.email ILIKE '%' || sqlc.arg(search) || '%'
              )
        )
    );

-- name: TouchConversationLastMessage :one
UPDATE conversations
SET
    last_message_at = $2,
    last_message_preview = $3
WHERE id = $1
RETURNING *;

-- name: UpdateWhatsAppContactProfile :one
UPDATE conversations
SET
    wa_contact_name = COALESCE(sqlc.narg(wa_contact_name), wa_contact_name),
    resident_id = COALESCE(sqlc.narg(resident_id), resident_id)
WHERE id = $1
RETURNING *;

-- name: MarkConversationRead :one
UPDATE conversation_participants
SET last_read_at = $3
WHERE conversation_id = $1 AND user_id = $2
RETURNING *;

-- name: EnsureConversationParticipant :exec
INSERT INTO conversation_participants (conversation_id, user_id, last_read_at)
VALUES ($1, $2, $3)
ON CONFLICT (conversation_id, user_id) DO NOTHING;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, sender_kind, body, wa_message_id, wa_status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMessageByID :one
SELECT *
FROM messages
WHERE id = $1;

-- name: GetMessageByWaMessageID :one
SELECT *
FROM messages
WHERE wa_message_id = $1;

-- name: UpdateMessageWaStatus :one
UPDATE messages
SET wa_status = $2
WHERE wa_message_id = $1
RETURNING *;

-- name: ListMessagesByConversation :many
SELECT *
FROM messages
WHERE conversation_id = $1
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: GetMaxOtherLastReadAt :one
SELECT MAX(last_read_at)::timestamptz AS last_read_at
FROM conversation_participants
WHERE conversation_id = sqlc.arg(conversation_id)
  AND user_id <> sqlc.arg(user_id);

-- name: CountMessagesByConversation :one
SELECT COUNT(*)::bigint
FROM messages
WHERE conversation_id = $1;

-- name: CountUnreadMessages :one
SELECT COUNT(*)::bigint
FROM messages m
INNER JOIN conversations c ON c.id = m.conversation_id
WHERE m.conversation_id = sqlc.arg(conversation_id)
  AND (
      (c.type = 'WHATSAPP' AND m.sender_kind = 'CONTACT')
      OR (
          c.type <> 'WHATSAPP'
          AND m.sender_kind = 'USER'
          AND m.sender_id IS DISTINCT FROM sqlc.arg(user_id)
      )
  )
  AND (
      NOT EXISTS (
          SELECT 1
          FROM conversation_participants me
          WHERE me.conversation_id = m.conversation_id
            AND me.user_id = sqlc.arg(user_id)
      )
      OR EXISTS (
          SELECT 1
          FROM conversation_participants me
          WHERE me.conversation_id = m.conversation_id
            AND me.user_id = sqlc.arg(user_id)
            AND (me.last_read_at IS NULL OR m.created_at > me.last_read_at)
      )
  );

-- name: ListChatContacts :many
SELECT id, email, name, role
FROM users
WHERE id <> sqlc.arg(user_id)
  AND (
      sqlc.arg(search)::text = ''
      OR name ILIKE '%' || sqlc.arg(search) || '%'
      OR email ILIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY name ASC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountChatContacts :one
SELECT COUNT(*)::bigint
FROM users
WHERE id <> sqlc.arg(user_id)
  AND (
      sqlc.arg(search)::text = ''
      OR name ILIKE '%' || sqlc.arg(search) || '%'
      OR email ILIKE '%' || sqlc.arg(search) || '%'
  );

-- name: GetChatUserByID :one
SELECT id, email, name, role
FROM users
WHERE id = $1;

-- name: FindResidentByPhoneCandidates :one
SELECT id, name, phone
FROM residents
WHERE phone = ANY(sqlc.arg(phones)::text[])
ORDER BY created_at ASC
LIMIT 1;
