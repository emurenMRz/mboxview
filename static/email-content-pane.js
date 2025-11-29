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

		// Create display container
		const displayContainer = document.createElement('div');
		displayContainer.style.display = 'flex';
		displayContainer.style.flexDirection = 'column';
		displayContainer.style.height = '100%';

		// Display Body with view toggle if both versions are available
		if (content.hasAlternate) {
			// Create toggle buttons
			const toggleDiv = document.createElement('div');
			toggleDiv.style.display = 'flex';
			toggleDiv.style.gap = '8px';
			toggleDiv.style.padding = '8px';
			toggleDiv.style.borderBottom = '1px solid #ccc';
			toggleDiv.style.backgroundColor = '#f5f5f5';

			const htmlBtn = document.createElement('button');
			htmlBtn.textContent = 'HTML View';
			htmlBtn.style.padding = '6px 12px';
			htmlBtn.style.cursor = 'pointer';
			htmlBtn.style.backgroundColor = '#007bff';
			htmlBtn.style.color = 'white';
			htmlBtn.style.border = 'none';
			htmlBtn.style.borderRadius = '4px';

			const textBtn = document.createElement('button');
			textBtn.textContent = 'Text View';
			textBtn.style.padding = '6px 12px';
			textBtn.style.cursor = 'pointer';
			textBtn.style.backgroundColor = '#ccc';
			textBtn.style.color = '#333';
			textBtn.style.border = 'none';
			textBtn.style.borderRadius = '4px';

			toggleDiv.appendChild(htmlBtn);
			toggleDiv.appendChild(textBtn);
			displayContainer.appendChild(toggleDiv);

			// Create content container
			const contentContainer = document.createElement('div');
			contentContainer.style.flex = '1';
			contentContainer.style.overflow = 'auto';
			contentContainer.id = 'email-body-container';
			displayContainer.appendChild(contentContainer);

			// HTML view (default)
			const displayHtmlView = () => {
				contentContainer.innerHTML = '';
				htmlBtn.style.backgroundColor = '#007bff';
				htmlBtn.style.color = 'white';
				textBtn.style.backgroundColor = '#ccc';
				textBtn.style.color = '#333';

				contentContainer.style.overflow = 'hidden';
				const iframe = document.createElement('iframe');
				iframe.setAttribute('sandbox', 'allow-same-origin');
				iframe.style.width = '100%';
				iframe.style.height = '100%';
				iframe.style.border = 'none';
				iframe.style.display = 'block';
				iframe.srcdoc = content.bodyHTML;
				contentContainer.appendChild(iframe);
			};

			// Text view
			const displayTextView = () => {
				contentContainer.innerHTML = '';
				htmlBtn.style.backgroundColor = '#ccc';
				htmlBtn.style.color = '#333';
				textBtn.style.backgroundColor = '#007bff';
				textBtn.style.color = 'white';

				contentContainer.style.overflow = 'auto';
				const pre = document.createElement('pre');
				pre.style.whiteSpace = 'pre-wrap';
				pre.style.wordWrap = 'break-word';
				pre.style.margin = '0';
				pre.style.padding = '8px';
				pre.style.boxSizing = 'border-box';
				pre.textContent = content.bodyText || 'No text version available.';
				contentContainer.appendChild(pre);
			};

			// Event listeners
			htmlBtn.addEventListener('click', displayHtmlView);
			textBtn.addEventListener('click', displayTextView);

			// Display Text by default
			displayTextView();
		} else if (content.bodyType === 'text/html') {
			// Single HTML version
			emailContent.style.overflow = 'hidden';

			const iframe = document.createElement('iframe');
			iframe.setAttribute('sandbox', 'allow-same-origin');
			iframe.style.width = '100%';
			iframe.style.height = '100%';
			iframe.style.border = 'none';
			iframe.style.display = 'block';
			iframe.srcdoc = content.bodyHTML;
			displayContainer.appendChild(iframe);
		} else {
			// Text-only version
			emailContent.style.overflow = 'auto';

			const pre = document.createElement('pre');
			pre.style.whiteSpace = 'pre-wrap';
			pre.style.wordWrap = 'break-word';
			pre.style.margin = '0';
			pre.style.padding = '8px';
			pre.style.boxSizing = 'border-box';
			pre.textContent = content.bodyText || 'No viewable content.';
			displayContainer.appendChild(pre);
		}

		emailContent.appendChild(displayContainer);

		// Display Attachments
		if (content.attachments && content.attachments.length > 0) {
			const attachmentsDiv = document.createElement('div');
			attachmentsDiv.style.marginTop = '20px';
			attachmentsDiv.style.padding = '8px';
			attachmentsDiv.style.borderTop = '1px solid #ccc';
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
