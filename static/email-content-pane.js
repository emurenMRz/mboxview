/**
 * Email Content Pane Handler
 * Manages loading and displaying email content and attachments
 */

async function loadEmail(emailContent, mailboxName, emailId) {
	emailContent.innerHTML = '<p>Loading content...</p>';
	try {
		const response = await fetch(`/api/mailboxes/${encodeURIComponent(mailboxName)}/emails/${emailId}`);
		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}
		const content = await response.json();

		emailContent.innerHTML = '';

		const displayContainer = document.createElement('div');
		displayContainer.className = 'email-display';

		if (content.hasAlternate) {
			const toggleDiv = document.createElement('div');
			toggleDiv.className = 'email-toggle-bar';

			const htmlBtn = document.createElement('button');
			htmlBtn.textContent = 'HTML View';
			htmlBtn.className = 'email-toggle-btn';

			const textBtn = document.createElement('button');
			textBtn.textContent = 'Text View';
			textBtn.className = 'email-toggle-btn active';

			toggleDiv.appendChild(htmlBtn);
			toggleDiv.appendChild(textBtn);
			displayContainer.appendChild(toggleDiv);

			const contentContainer = document.createElement('div');
			contentContainer.className = 'email-body-container email-text';
			contentContainer.id = 'email-body-container';
			displayContainer.appendChild(contentContainer);

			const displayHtmlView = () => {
				contentContainer.innerHTML = '';
				contentContainer.className = 'email-body-container email-iframe';
				htmlBtn.classList.add('active');
				textBtn.classList.remove('active');

				const iframe = document.createElement('iframe');
				iframe.setAttribute('sandbox', 'allow-same-origin');
				iframe.className = 'email-body-iframe';
				iframe.srcdoc = content.bodyHTML;
				contentContainer.appendChild(iframe);
			};

			const displayTextView = () => {
				contentContainer.innerHTML = '';
				contentContainer.className = 'email-body-container email-text';
				htmlBtn.classList.remove('active');
				textBtn.classList.add('active');

				const pre = document.createElement('pre');
				pre.className = 'email-body-text';
				pre.textContent = content.bodyText || 'No text version available.';
				contentContainer.appendChild(pre);
			};

			htmlBtn.addEventListener('click', displayHtmlView);
			textBtn.addEventListener('click', displayTextView);

			displayTextView();
		} else if (content.bodyType === 'text/html') {
			emailContent.style.overflow = 'hidden';

			const iframe = document.createElement('iframe');
			iframe.setAttribute('sandbox', 'allow-same-origin');
			iframe.className = 'email-body-iframe';
			iframe.srcdoc = content.bodyHTML;
			displayContainer.appendChild(iframe);
		} else {
			emailContent.style.overflow = 'auto';

			const pre = document.createElement('pre');
			pre.className = 'email-body-text';
			pre.textContent = content.bodyText || 'No viewable content.';
			displayContainer.appendChild(pre);
		}

		emailContent.appendChild(displayContainer);

		if (content.attachments && content.attachments.length > 0) {
			const attachmentsDiv = document.createElement('div');
			attachmentsDiv.className = 'email-attachments';
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
