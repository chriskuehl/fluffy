function changeStyleTo(style) {
    document.getElementById('highlightContainer').className = 'highlight-' + style;
    if (hasLocalStorage) {
        localStorage.setItem(preferredStyleVar, style);
    }
}

var preferredStyle = null;
if (hasLocalStorage) {
    preferredStyle = localStorage.getItem(preferredStyleVar);
    if (preferredStyle !== null) {
        changeStyleTo(preferredStyle);
    }
}
