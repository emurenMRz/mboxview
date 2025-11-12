document.addEventListener('DOMContentLoaded', () => {
    const folderList = document.getElementById('folder-list');
    const emailListBody = document.querySelector('#email-list tbody');
    const emailContent = document.getElementById('email-content');

    let currentMailbox = null;

    async function loadEmailContent(mailboxName, emailId) {
        emailContent.innerHTML = '<p>Loading content...</p>';
        try {
            const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails/${emailId}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const content = await response.json();

            emailContent.innerHTML = ''; // Clear loading message

            // Display Body
            if (content.bodyType === 'text/html') {
                // For HTML content we embed in an iframe. Prevent the parent pane
                // from showing its own scrollbar to avoid duplicated scrollbars
                // (iframe will host its own scrolling if needed).
                emailContent.style.overflow = 'hidden';

                const iframe = document.createElement('iframe');
                iframe.setAttribute('sandbox', 'allow-same-origin');
                iframe.style.width = '100%';
                iframe.style.height = '100%';
                iframe.style.border = 'none';
                iframe.style.display = 'block';
                iframe.srcdoc = content.body;
                emailContent.appendChild(iframe);
            } else {
                // For plain text, let the parent pane (emailContent) scroll.
                emailContent.style.overflow = 'auto';

                const pre = document.createElement('pre');
                pre.style.whiteSpace = 'pre-wrap';
                pre.style.wordWrap = 'break-word';
                pre.style.margin = '0';
                pre.style.padding = '8px';
                pre.style.boxSizing = 'border-box';
                pre.textContent = content.body || 'No viewable content.';
                emailContent.appendChild(pre);
            }

            // Display Attachments
            if (content.attachments && content.attachments.length > 0) {
                const attachmentsDiv = document.createElement('div');
                attachmentsDiv.style.marginTop = '20px';
                attachmentsDiv.innerHTML = '<strong>Attachments:</strong>';
                const ul = document.createElement('ul');
                content.attachments.forEach(name => {
                    const li = document.createElement('li');
                    li.textContent = name;
                    ul.appendChild(li);
                });
                attachmentsDiv.appendChild(ul);
                emailContent.appendChild(attachmentsDiv);
            }

        } catch (error) {
            console.error(`Failed to load email content for ID ${emailId}:`, error);
            emailContent.innerHTML = '<p>Error loading email content.</p>';
        }
    }

    async function loadEmails(mailboxName) {
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
                    loadEmailContent(mailboxName, email.id);
                    document.querySelectorAll('#email-list tbody tr').forEach(item => item.classList.remove('selected'));
                    row.classList.add('selected');
                });

                emailListBody.appendChild(row);
            });

        } catch (error) {
            console.error(`Failed to load emails for ${mailboxName}:`, error);
            emailListBody.innerHTML = '<tr><td colspan="3">Error loading emails.</td></tr>';
        }
    }

    // Load initial folder list
    async function loadFolders() {
        try {
            const response = await fetch('/api/mailboxes');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const folders = await response.json();
            
            folderList.innerHTML = ''; // Clear existing list
            folders.forEach(folder => {
                const li = document.createElement('li');
                li.textContent = folder;
                li.dataset.mailbox = folder;
                li.addEventListener('click', () => {
                    // Handle folder selection
                    if (currentMailbox !== folder) {
                        currentMailbox = folder;
                        // Update selected visual state
                        document.querySelectorAll('#folder-list li').forEach(item => item.classList.remove('selected'));
                        li.classList.add('selected');
                        
                        // Load emails for the selected folder
                        loadEmails(folder);

                        emailContent.innerHTML = '<p>Select an email to view its content.</p>'; // Reset content
                    }
                });
                folderList.appendChild(li);
            });
        } catch (error) {
            console.error('Failed to load folders:', error);
            folderList.innerHTML = '<li>Error loading folders.</li>';
        }
    }

    loadFolders();
});
