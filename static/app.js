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

    // Pane resizer logic: allow dragging the divider to resize #folders-pane
    resizeFolderPane();

    // Vertical resizer for emails/content panes
    resizeEmailsPane();
});

function resizeFolderPane() {
    const resizer = document.getElementById('pane-resizer');
    const foldersPane = document.getElementById('folders-pane');
    const MIN_WIDTH = 200; // px
    const MAX_WIDTH = 300; // px
    let isDragging = false;

    let startX = 0;
    let startWidth = 0;
    function onPointerDown(e) {
        if (e.isPrimary === false) return;
        isDragging = true;
        document.body.style.cursor = 'col-resize';
        startX = e.clientX;
        startWidth = foldersPane.getBoundingClientRect().width;
        try {
            resizer.setPointerCapture && resizer.setPointerCapture(e.pointerId);
        } catch (err) {
        }
        e.preventDefault();
    }

    function onPointerUp(e) {
        if (!isDragging) return;
        isDragging = false;
        document.body.style.cursor = '';
        try {
            resizer.releasePointerCapture && resizer.releasePointerCapture(e.pointerId);
        } catch (err) {
            // ignore
        }
    }

    function onPointerMove(e) {
        if (!isDragging) return;
        const clientX = e.clientX;
        let delta = clientX - startX;
        let newWidth = Math.round(startWidth + delta);
        if (newWidth < MIN_WIDTH) newWidth = MIN_WIDTH;
        if (newWidth > MAX_WIDTH) newWidth = MAX_WIDTH;
        foldersPane.style.flex = '0 0 ' + newWidth + 'px';
        e.preventDefault();
    }

    resizer.addEventListener('pointerdown', onPointerDown);
    window.addEventListener('pointerup', onPointerUp);
    window.addEventListener('pointermove', onPointerMove);
}

function resizeEmailsPane() {
    const resizer = document.getElementById('vertical-resizer');
    const emailsPane = document.getElementById('emails-pane');
    const MIN_HEIGHT = 120; // px
    let isDragging = false;

    let startY = 0;
    let startHeight = 0;
    function onPointerDown(e) {
        if (e.isPrimary === false) return;
        isDragging = true;
        document.body.style.cursor = 'row-resize';
        startY = e.clientY;
        startHeight = emailsPane.getBoundingClientRect().height;
        try {
            resizer.setPointerCapture && resizer.setPointerCapture(e.pointerId);
        } catch (err) {
        }
        e.preventDefault();
    }

    function onPointerUp(e) {
        if (!isDragging) return;
        isDragging = false;
        document.body.style.cursor = '';
        try {
            resizer.releasePointerCapture && resizer.releasePointerCapture(e.pointerId);
        } catch (err) {
        }
    }

    function onPointerMove(e) {
        if (!isDragging) return;
        const clientY = e.clientY;
        let delta = clientY - startY;
        let newHeight = Math.round(startHeight + delta);
        const rect = document.getElementById('right-container').getBoundingClientRect();
        const maxAllowed = Math.max(rect.height - MIN_HEIGHT, MIN_HEIGHT);
        if (newHeight < MIN_HEIGHT) newHeight = MIN_HEIGHT;
        if (newHeight > maxAllowed) newHeight = maxAllowed;
        emailsPane.style.flex = '0 0 ' + newHeight + 'px';
        e.preventDefault();
    }

    resizer.addEventListener('pointerdown', onPointerDown);
    window.addEventListener('pointerup', onPointerUp);
    window.addEventListener('pointermove', onPointerMove);
}
