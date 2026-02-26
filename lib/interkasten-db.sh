#!/usr/bin/env bash
# Shared SQLite access to interkasten's entity_map.
#
# Contract:
# - interkasten owns sync columns (entity_key, entity_type, notion_id, local_path, base_content_hash, ...)
# - intertree owns hierarchy columns (parent_id, tags, doc_tier)
# - Both may read shared columns (entity_key, entity_type, local_path, notion_id)
#
# Usage: source this file and call ib_kasten_* functions.
# Fail-open: all functions return safe defaults if sqlite3 or DB missing.

[[ -n "${_INTERBASE_KASTEN_DB_LOADED:-}" ]] && return 0
_INTERBASE_KASTEN_DB_LOADED=1

# Default DB path
_IB_KASTEN_DB="${INTERKASTEN_DB:-${HOME}/.interkasten/state.db}"

# --- Guards ---

ib_kasten_has_db() {
    command -v sqlite3 &>/dev/null && [[ -f "$_IB_KASTEN_DB" ]]
}

# --- Read operations ---

# List all registered projects as JSON array
# Output: [{"id":1,"local_path":"/path","parent_id":null,"tags":"[\"tag\"]","doc_tier":"Product"}]
ib_kasten_list_projects() {
    ib_kasten_has_db || { echo "[]"; return 0; }
    sqlite3 -json "$_IB_KASTEN_DB" \
        "SELECT id, local_path, parent_id, tags, doc_tier FROM entity_map WHERE entity_type = 'project' ORDER BY local_path" \
        2>/dev/null || echo "[]"
}

# Get a project by local path
# Output: JSON object or empty string
ib_kasten_get_project() {
    local path="$1"
    ib_kasten_has_db || { echo ""; return 0; }
    sqlite3 -json "$_IB_KASTEN_DB" \
        "SELECT id, local_path, parent_id, tags, doc_tier, notion_id FROM entity_map WHERE entity_type = 'project' AND local_path = '$path' LIMIT 1" \
        2>/dev/null | jq -r '.[0] // empty' 2>/dev/null || echo ""
}

# Get children of a project by parent_id
# Output: JSON array
ib_kasten_get_children() {
    local parent_id="$1"
    ib_kasten_has_db || { echo "[]"; return 0; }
    sqlite3 -json "$_IB_KASTEN_DB" \
        "SELECT id, local_path, tags, doc_tier FROM entity_map WHERE parent_id = $parent_id ORDER BY local_path" \
        2>/dev/null || echo "[]"
}

# --- Hierarchy write operations (intertree-owned columns) ---

# Set parent_id for a project
ib_kasten_set_parent() {
    local project_id="$1" parent_id="${2:-NULL}"
    ib_kasten_has_db || return 1
    sqlite3 "$_IB_KASTEN_DB" \
        "UPDATE entity_map SET parent_id = $parent_id WHERE id = $project_id" \
        2>/dev/null
}

# Set tags for a project (JSON array string)
ib_kasten_set_tags() {
    local project_id="$1" tags_json="$2"
    ib_kasten_has_db || return 1
    sqlite3 "$_IB_KASTEN_DB" \
        "UPDATE entity_map SET tags = '$tags_json' WHERE id = $project_id" \
        2>/dev/null
}

# Set doc_tier for a project
ib_kasten_set_doc_tier() {
    local project_id="$1" tier="$2"
    ib_kasten_has_db || return 1
    sqlite3 "$_IB_KASTEN_DB" \
        "UPDATE entity_map SET doc_tier = '$tier' WHERE id = $project_id" \
        2>/dev/null
}
