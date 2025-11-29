/**
 * Pane Resizer Logic
 * Handles dragging and resizing of panes
 */

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
