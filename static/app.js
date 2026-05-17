/**
 * Main Application Entry Point
 * Coordinates initialization and pane communication
 */

/* ── Toast Notification ── */
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    container.appendChild(toast);
    setTimeout(() => {
        toast.classList.add('toast-out');
        toast.addEventListener('animationend', () => toast.remove());
    }, 2500);
}

/* ── Confirm Modal ── */
function showConfirm(message) {
    return new Promise((resolve) => {
        const overlay = document.getElementById('confirm-modal-overlay');
        const msgEl = document.getElementById('confirm-modal-message');
        const okBtn = document.getElementById('confirm-ok-btn');
        const cancelBtn = document.getElementById('confirm-cancel-btn');

        msgEl.textContent = message;
        overlay.classList.add('modal-show');

        const cleanup = () => {
            okBtn.removeEventListener('click', onOk);
            cancelBtn.removeEventListener('click', onCancel);
            overlay.removeEventListener('click', onOverlay);
            overlay.classList.remove('modal-show');
        };

        const onOk = () => { cleanup(); resolve(true); };
        const onCancel = () => { cleanup(); resolve(false); };
        const onOverlay = (e) => { if (e.target === overlay) onCancel(); };

        okBtn.addEventListener('click', onOk);
        cancelBtn.addEventListener('click', onCancel);
        overlay.addEventListener('click', onOverlay);
    });
}

/* ── Status Bar ── */
function updateStatusBar() {
    const mailboxEl = document.getElementById('status-mailbox');
    const countEl = document.getElementById('status-count');
    const selectedEl = document.getElementById('status-selected');

    if (mailboxEl) mailboxEl.textContent = window._currentMailbox || '—';
    if (countEl) countEl.textContent = window._emailCount ? `${window._emailCount}件` : '—';

    const checked = document.querySelectorAll('.email-checkbox:checked').length;
    if (selectedEl) selectedEl.textContent = checked ? `${checked}件選択中` : '';
}

document.addEventListener('DOMContentLoaded', () => {
    const folderList = document.getElementById('folder-list');
    const emailListBody = document.querySelector('#email-list tbody');
    const emailContent = document.getElementById('email-content');

    let currentMailbox = null;
    window._currentMailbox = null;
    window._emailCount = null;

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
            window._currentMailbox = folder;
            updateStatusBar();
            // Load emails for the selected folder
            loadEmails(emailListBody, folder, (mailboxName, emailId) => {
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
    updateStatusBar();
}

async function deleteBatchEmails(mailboxName) {
    const checkboxes = Array.from(document.querySelectorAll('.email-checkbox:checked'));

    if (checkboxes.length === 0) {
        showToast('削除するメールを選択してください。', 'info');
        return;
    }

    const emailIds = checkboxes.map(cb => parseInt(cb.dataset.emailId));

    const confirmed = await showConfirm(`${emailIds.length}件のメールを削除しますか？`);
    if (!confirmed) return;

    try {
        const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails/delete-batch`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ids: emailIds })
        });

        if (!response.ok)
            throw new Error(`HTTP error! status: ${response.status}`);

        const result = await response.json();

        checkboxes.forEach(cb => {
            const row = cb.closest('tr');
            if (row) row.remove();
        });

        if (typeof allEmails !== 'undefined' && Array.isArray(allEmails))
            allEmails = allEmails.filter(email => !emailIds.some(deletedId => Number(email.id) === Number(deletedId)));

        const container = document.getElementById('batch-delete-container');
        container.style.display = 'none';

        window._emailCount = (window._emailCount || 0) - result.deleted;
        updateStatusBar();

        if (result.failed > 0)
            showToast(`${result.deleted}件削除、${result.failed}件失敗`, 'error');
        else
            showToast(`${result.deleted}件削除しました`, 'success');
    } catch (error) {
        console.error('Failed to delete batch emails:', error);
        showToast('メールの削除に失敗しました。', 'error');
    }
}
