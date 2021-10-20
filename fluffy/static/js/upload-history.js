(() => {
    const container = document.getElementById('history');
    const entryTemplate = document.getElementById('template-entry').content.firstElementChild;
    const fileListEntryTemplate = document.getElementById('template-file-list-entry').content.firstElementChild;

    if (uploadHistory.enabled()) {
        uploadHistory.getHistory().reverse().forEach(upload => {
            const entry = entryTemplate.cloneNode(true);
            entry.classList.add(upload.fileDetails ? 'file-upload' : 'paste');

            entry.querySelector('.fill').setAttribute('href', upload.url);
            entry.querySelector('.time').innerText = getTimeAgo(upload.time);

            const fileList = entry.querySelector('.file-list');
            const title = entry.querySelector('.title');
            if (upload.fileDetails) {
                const ud = upload.fileDetails;
                title.innerText = `${ud.length} file${s(ud.length)} uploaded,`;

                ud.forEach(file => {
                    const fileEntry = fileListEntryTemplate.cloneNode(true);
                    const extension = file.filename.split('.').pop();
                    fileEntry.querySelector('.icon').setAttribute('src', icons[extension] || icons['unknown']);
                    fileEntry.querySelector('.name').innerText = file.filename;
                    fileList.append(fileEntry);
                });
            } else {
                const pd = upload.pasteDetails;
                title.innerText = `${pd.num_lines} line${s(pd.num_lines)} of ${pd.language_title},`;
                fileList.remove();
            }

            container.append(entry);
        });
    }
})();
