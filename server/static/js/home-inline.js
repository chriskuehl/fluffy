// determine whether to use basic or advanced upload form
if (typeof FileReader != "undefined" &&
    "draggable" in document.createElement("span")) {
    document.getElementById("basicUpload").style.display = "none";
    document.getElementById("advancedUpload").style.display = "block";
} else {
    document.getElementById("oldBrowserMessage").style.display = "block";
}

(() => {
    const template = document.getElementById('recent-uploads-item-tmpl').content.firstElementChild;
    const grid = document.getElementById('recent-uploads-grid');

    const addEntry = (url, image, title, subtitle) => {
        const entry = template.cloneNode(true);
        entry.setAttribute('href', url);
        entry.querySelector('.recent-uploads-image').setAttribute('src', image);
        entry.querySelector('.recent-uploads-text-title').innerText = title;
        entry.querySelector('.recent-uploads-text-subtitle').innerText = subtitle;
        grid.prepend(entry);
    };

    const titlesForEntry = (entry) => {
        let title;

        if (entry.pasteDetails) {
            const lines = entry.pasteDetails.num_lines;
            title = `${lines} line${s(lines)} of ${entry.pasteDetails.language_title}`;
        } else {
            const extensions = new Set(
                entry.fileDetails
                    .map(f => f.filename.split('.').pop().toLowerCase())
                    // Limit to extensions we know about to avoid things like "blah.x86_64"
                    // being shown as "1 x86_64 file"
                    .filter(extension => icons[extension] !== undefined)
            );
            const fileType = extensions.size === 1 ? ` ${extensions.values().next().value.toUpperCase()} ` : '';
            const count = entry.fileDetails.length;
            title = `${count} ${fileType}file${s(count)}`;
        }

        return {
            title,
            subtitle: getTimeAgo(entry.time),
        }
    };

    const iconForEntry = (entry) => {
        if (entry.pasteDetails) {
            return icons['paste-generic'];
        } else {
            // Use the icon from the largest file uploaded that has a known extension.
            const candidateIcons = entry.fileDetails
                .map(({ filename, bytes }) => [icons[filename.split('.').pop()], bytes])
                .filter(([filename, _]) => filename);
            if (candidateIcons.length == 0) {
                return icons['unknown'];
            } else {
                // Is there a graceful way to do max() with a key function in JS?
                const max = Math.max(...candidateIcons.map(([_, bytes]) => bytes));
                return candidateIcons.filter(([_, bytes]) => bytes == max)[0][0];
            }
        }
    };

    if (uploadHistory.enabled()) {
        const entries = uploadHistory.getHistory().slice(-3);
        if (entries.length > 0) {
            entries.forEach(entry => {
                const { title, subtitle } = titlesForEntry(entry);
                addEntry(entry.url, iconForEntry(entry), title, subtitle);
            });
            document.getElementById('recent-uploads').classList.remove('hidden');
        }
    }
})();
