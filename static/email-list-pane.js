/**
 * Email List Pane Handler
 * Manages loading and displaying email list
 */

async function loadEmails(emailListBody, mailboxName, onEmailSelected) {
    emailListBody.innerHTML = '<tr><td colspan="3">Loading...</td></tr>';
    try {
        const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const emails = await response.json();

        // Clear existing selected state
        document.querySelectorAll('#email-list tbody tr.selected').forEach(row => {
            row.classList.remove('selected');
        });

        emailListBody.innerHTML = ''; // Clear loading message
        if (!emails || emails.length === 0) {
            emailListBody.innerHTML = '<tr><td colspan="3">No emails in this mailbox.</td></tr>';
            return;
        }

        emails.forEach(email => {
            const row = document.createElement('tr');
            row.dataset.emailId = email.id;
            // 新着メール（StatusにNまたはUが含まれる）は太字
            if (email.status.includes('N') || email.status.includes('U'))
                row.classList.add('new-mail-row');

            const dateCell = document.createElement('td');
            dateCell.textContent = new Date(email.date).toLocaleString();

            const fromCell = document.createElement('td');
            fromCell.textContent = email.from;

            const subjectCell = document.createElement('td');
            subjectCell.textContent = email.subject;

            row.appendChild(dateCell);
            row.appendChild(fromCell);
            row.appendChild(subjectCell);

            // Add delete button
            const deleteCell = document.createElement('td');
            const deleteButton = document.createElement('button');
            deleteButton.textContent = '✖';
            deleteButton.className = 'delete-email';
            deleteButton.addEventListener('click', (e) => {
                e.stopPropagation(); // Prevent row click
                deleteEmail(mailboxName, email.id, row);
            });
            deleteCell.appendChild(deleteButton);
            row.appendChild(deleteCell);

            row.addEventListener('click', () => {
                // Clear selection from all rows
                document.querySelectorAll('#email-list tbody tr').forEach(r => {
                    r.classList.remove('selected');
                });
                // Mark clicked row as selected
                row.classList.add('selected');
                onEmailSelected(mailboxName, email.id);
            });
            
            // Also mark as read when double-clicked
            row.addEventListener('dblclick', () => {
                markEmailAsRead(mailboxName, email.id);
            });

            emailListBody.appendChild(row);
        });

    } catch (error) {
        console.error(`Failed to load emails for ${mailboxName}:`, error);
        emailListBody.innerHTML = '<tr><td colspan="3">Error loading emails.</td></tr>';
    }
}

async function markEmailAsRead(mailboxName, emailId) {
    try {
        const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails/${emailId}/read`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        });
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        // Update UI for the specific email row only
        const row = document.querySelector('tr[data-email-id="' + emailId + '"]');
        if (row) {
            row.classList.remove('new-mail-row');
        }
    } catch (error) {
        console.error(`Failed to mark email ${emailId} as read:`, error);
    }
}

async function deleteEmail(mailboxName, emailId, rowElement) {
    if (!confirm('本当にこのメールを削除しますか？')) {
        return;
    }
    try {
        const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails/${emailId}`, {
            method: 'DELETE',
        });
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        // Remove the row from UI
        rowElement.remove();
    } catch (error) {
        console.error(`Failed to delete email ${emailId}:`, error);
        alert('メールの削除に失敗しました。');
    }
}
