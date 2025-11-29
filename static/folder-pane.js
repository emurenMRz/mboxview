/**
 * Folder Pane Handler
 * Manages loading and selecting mailbox folders
 */

async function loadFolders(folderList, onFolderSelected) {
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
                // Update selected visual state
                document.querySelectorAll('#folder-list li').forEach(item => item.classList.remove('selected'));
                li.classList.add('selected');

                // Notify parent of folder selection
                onFolderSelected(folder);
            });
            folderList.appendChild(li);
        });
    } catch (error) {
        console.error('Failed to load folders:', error);
        folderList.innerHTML = '<li>Error loading folders.</li>';
    }
}
