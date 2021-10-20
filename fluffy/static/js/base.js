// https://developer.mozilla.org/en-US/docs/Web/API/Web_Storage_API/Using_the_Web_Storage_API#feature-detecting_localstorage
const hasLocalStorage = ((type) => {
    var storage;
    try {
        storage = window[type];
        var x = '__storage_test__';
        storage.setItem(x, x);
        storage.removeItem(x);
        return true;
    }
    catch(e) {
        return e instanceof DOMException && (
            // everything except Firefox
            e.code === 22 ||
            // Firefox
            e.code === 1014 ||
            // test name field too, because code might not be present
            // everything except Firefox
            e.name === 'QuotaExceededError' ||
            // Firefox
            e.name === 'NS_ERROR_DOM_QUOTA_REACHED') &&
            // acknowledge QuotaExceededError only if there's something already stored
            (storage && storage.length !== 0);
    }
})('localStorage');

const s = (n => (n == 1 ? '' : 's'));

// Adapted from https://gist.github.com/pomber/6195066a9258d1fb93bb59c206345b38
const getTimeAgo = (date) => {
    const MINUTE = 60,
        HOUR = MINUTE * 60,
        DAY = HOUR * 24,
        MONTH = DAY * 30,
        YEAR = MONTH * 12;
    const secondsAgo = Math.round((+new Date() - date) / 1000);

    let result;
    if (secondsAgo < MINUTE) {
        result = [secondsAgo, 'second'];
    } else if (secondsAgo < HOUR) {
        result = [Math.floor(secondsAgo / MINUTE), 'minute'];
    } else if (secondsAgo < DAY) {
        result = [Math.floor(secondsAgo / HOUR), 'hour'];
    } else if (secondsAgo < MONTH) {
        result = [Math.floor(secondsAgo / DAY), 'day'];
    } else if (secondsAgo < YEAR) {
        result = [Math.floor(secondsAgo / MONTH), 'month'];
    } else {
        result = [Math.floor(secondsAgo / YEAR), 'year'];
    }

    return `${result[0]} ${result[1]}${s(result[0])} ago`;
};

const uploadHistory = {
    enabled: () => hasLocalStorage,

    getHistory: () => {
        if (localStorage.uploadHistory) {
            return JSON.parse(localStorage.uploadHistory).map(entry => ({
                url: entry.url,
                time: new Date(entry.time),
                fileDetails: entry.fileDetails,
                pasteDetails: entry.pasteDetails,
            }));
        } else {
            return [];
        }
    },

    addItemToHistory: ({ url, time, fileDetails, pasteDetails }) => {
        const uploadHistory = localStorage.uploadHistory ? JSON.parse(localStorage.uploadHistory) : [];
        const entry = {
            url,
            time: time.getTime(),
            fileDetails,
            pasteDetails,
            version: 1,
        };
        uploadHistory.push(entry);
        localStorage.uploadHistory = JSON.stringify(uploadHistory);
    },

    clearHistory: () => {
        delete localStorage.uploadHistory;
    },
};
