const debouncedQuery = _.debounce(async (inputValue, processResponse, clearComponent) => {
    console.log(inputValue);

    try {
      if (inputValue.length < 3) {
        clearComponent();
        return;
      }

      const url = `${apiUrl}?query=${encodeURIComponent(inputValue)}`;

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      processResponse(data);

    } catch (error) {
      console.error('Error fetching data:', error);
    }
}, 300);

async function queryApiOnInput(inputElement, apiUrl, processResponse, clearComponent) {
  inputElement.addEventListener('input', (event) => debouncedQuery(event.target.value, processResponse, clearComponent));
}

const inputElement = document.getElementById('input');
const apiUrl = 'http://localhost:1234/data';

function handleApiResponse(data) {
  console.log('API response:', data);

  const resultsDiv = document.getElementById('apiResponse');
  if (resultsDiv){
    resultsDiv.innerHTML = '';

    const counter = data.counter;
    if (Array.isArray(data.items)){
      data.items.forEach(item => {
        const resultItem = document.createElement('div');

        resultItem.textContent = JSON.stringify(item);
        resultsDiv.appendChild(resultItem);
      });
    } else {
      resultsDiv.textContent = JSON.stringify(data);
    }
  }
}

function clearResponseComponent() {
  const resultsDiv = document.getElementById('apiResponse');
  if (resultsDiv){
    resultsDiv.innerHTML = '';
  }
}

// Start querying the API when the user types in the input.
queryApiOnInput(inputElement, apiUrl, handleApiResponse, clearResponseComponent);
