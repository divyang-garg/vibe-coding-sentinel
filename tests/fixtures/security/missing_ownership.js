// SEC-001: Missing Resource Ownership Check
// This file contains functions that access resources without ownership verification

// VULNERABLE: Access resource without ownership check
function getUserPosts(userId) {
    // Missing check: req.user.id === userId
    return db.query("SELECT * FROM posts WHERE userId = ?", [userId]);
}

// VULNERABLE: Update resource without ownership verification
function updatePost(postId, data) {
    // Missing ownership check before update
    return db.query("UPDATE posts SET title = ? WHERE id = ?", [data.title, postId]);
}

// VULNERABLE: Delete resource without ownership check
function deletePost(postId) {
    // Missing: Check if req.user.id === post.userId
    return db.query("DELETE FROM posts WHERE id = ?", [postId]);
}

// SAFE: Resource access with ownership check (should not be flagged)
function getUserPostsSafe(req, res) {
    const userId = req.params.id;
    // Ownership check present
    if (req.user.id !== userId && req.user.role !== 'admin') {
        return res.status(403).json({ error: 'Forbidden' });
    }
    return db.query("SELECT * FROM posts WHERE userId = ?", [userId])
        .then(posts => res.json(posts));
}

