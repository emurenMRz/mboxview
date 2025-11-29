/**
 * Email Content Pane Handler
 * Manages loading and displaying email content and attachments
 */

async function loadEmail(emailContent, mailboxName, emailId) {
    // Load email content when an email is selected
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
