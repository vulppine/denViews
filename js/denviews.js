/* eslint-env browser */

function queryViews () {
  const currentURL = new URL(document.URL)
  currentURL.host = document.querySelector('#denviews').dataset.denviewsHost

  fetch(currentURL.toString())
    .then((response) => {
      if (!response.ok) {
        throw new Error('could not retrieve views for page')
      }

      return response.json()
    })
    .then((data) => {
      const viewCount = document.querySelector('#denviews-viewcount')
      const hitCount = document.querySelector('#denviews-hitcount')

      if (viewCount !== null) {
        document.querySelector('#denviews-viewcount').innerText = data.views
      }

      if (hitCount !== null) {
        document.querySelector('#denviews-hitcount').innerText = data.hits
      }
    })
}

queryViews()
