/**
 * Main Application Entry Point
 * Coordinates initialization and pane communication
 */

document.addEventListener('DOMContentLoaded', () => {
    const folderList = document.getElementById('folder-list');
    const emailListBody = document.querySelector('#email-list tbody');
    const emailContent = document.getElementById('email-content');

    let currentMailbox = null;

    // Create batch delete button and insert into DOM
    const emailFilter = document.getElementById('email-filter');
    const batchDeleteContainer = document.createElement('div');
    batchDeleteContainer.id = 'batch-delete-container';
    batchDeleteContainer.style.display = 'none';

    const batchDeleteButton = document.createElement('button');
    batchDeleteButton.id = 'batch-delete-btn';
    batchDeleteButton.textContent = '選択したメールを削除';
    batchDeleteButton.addEventListener('click', () => {
        deleteBatchEmails(currentMailbox);
    });
    batchDeleteContainer.appendChild(batchDeleteButton);
    emailFilter.appendChild(batchDeleteContainer);

    // Initialize folder pane with callback
    loadFolders(folderList, (folder) => {
        if (currentMailbox !== folder) {
            currentMailbox = folder;
            // Load emails for the selected folder
            loadEmails(emailListBody, folder, (mailboxName, emailId) => {
                // When email is selected: load content and mark as read
                loadEmail(emailContent, mailboxName, emailId);
                markEmailAsRead(mailboxName, emailId);
            });
            emailContent.innerHTML = '<p>Select an email to view its content.</p>';
        }
    });

    // Initialize pane resizers
    resizeFolderPane();
    resizeEmailsPane();
});

function updateBatchDeleteButton() {
    const checkboxes = document.querySelectorAll('.email-checkbox:checked');
    const container = document.getElementById('batch-delete-container');
    if (checkboxes.length > 0) {
        container.style.display = 'block';
    } else {
        container.style.display = 'none';
    }
}

async function deleteBatchEmails(mailboxName) {
    const checkboxes = Array.from(document.querySelectorAll('.email-checkbox:checked'));

    if (checkboxes.length === 0) {
        alert('削除するメールを選択してください。');
        return;
    }

    const emailIds = checkboxes.map(cb => parseInt(cb.dataset.emailId));

    if (!confirm(`${emailIds.length}件のメールを削除しますか？`))
        return;

    try {
        const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails/delete-batch`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ ids: emailIds })
        });

        if (!response.ok)
            throw new Error(`HTTP error! status: ${response.status}`);

        const result = await response.json();

        // Remove rows from UI
        checkboxes.forEach(cb => {
            const row = cb.closest('tr');
            if (row) row.remove();
        });

        // Update allEmails by filtering out deleted emails
        if (typeof allEmails !== 'undefined' && Array.isArray(allEmails))
            allEmails = allEmails.filter(email => !emailIds.some(deletedId => Number(email.id) === Number(deletedId)));

        // Reset button and show summary
        const container = document.getElementById('batch-delete-container');
        container.style.display = 'none';
        alert(`${result.deleted}件削除、${result.failed}件失敗しました。`);
    } catch (error) {
        console.error('Failed to delete batch emails:', error);
        alert('メールの削除に失敗しました。');
    }
}

