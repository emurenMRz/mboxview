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

            row.addEventListener('click', () => {
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
