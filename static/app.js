/**
 * Main Application Entry Point
 * Coordinates initialization and pane communication
 */

document.addEventListener('DOMContentLoaded', () => {
    const folderList = document.getElementById('folder-list');
    const emailListBody = document.querySelector('#email-list tbody');
    const emailContent = document.getElementById('email-content');

    let currentMailbox = null;

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
