// Store chart instance
let stockChart = null;
let companyNames = {};

async function validateCSVData(csvData, fileName) {
    const lines = csvData.split('\n');
    if (lines.length < 2) {
        throw new Error(`${fileName} is empty or contains only headers`);
    }

    const headers = lines[0].split(',');
    const expectedHeaders = ['Date', 'Open', 'High', 'Low', 'Close', 'Change', 'Change%', 'Volume', 'T.Shares', 'Trades'];
    
    // Check headers
    if (!expectedHeaders.every((header, index) => headers[index] === header)) {
        throw new Error(`Invalid headers in ${fileName}. Expected: ${expectedHeaders.join(', ')}`);
    }

    // Check data format in first row
    const firstDataRow = lines[1].split(',');
    if (firstDataRow.length !== expectedHeaders.length) {
        throw new Error(`Invalid data format in ${fileName}. Expected ${expectedHeaders.length} columns`);
    }

    // Validate date format (DD/MM/YYYY)
    const dateRegex = /^\d{2}\/\d{2}\/\d{4}$/;
    if (!dateRegex.test(firstDataRow[0])) {
        throw new Error(`Invalid date format in ${fileName}. Expected DD/MM/YYYY`);
    }

    return true;
}

async function loadTickers() {
    try {
        const response = await fetch('/tickers/TICKERS.csv');
        if (!response.ok) {
            throw new Error(`Failed to fetch TICKERS.csv: ${response.statusText}`);
        }

        const data = await response.text();
        if (!data.trim()) {
            throw new Error('TICKERS.csv is empty');
        }

        const rows = data.split('\n').slice(1);
        if (rows.length === 0) {
            throw new Error('No ticker data found in TICKERS.csv');
        }

        const select = document.getElementById('tickerSelect');
        select.innerHTML = '<option value="">Select Ticker</option>';
        
        rows.forEach(row => {
            const columns = row.split(',');
            if (columns.length < 2) {
                console.warn('Invalid row format:', row);
                return;
            }

            const ticker = columns[0];
            const description = columns[1];
            
            if (ticker && ticker.trim()) {
                companyNames[ticker.trim()] = description.trim();
                const option = document.createElement('option');
                option.value = ticker.trim();
                option.text = ticker.trim();
                select.appendChild(option);
            }
        });

        if (select.options.length <= 1) {
            throw new Error('No valid tickers found in TICKERS.csv');
        }
    } catch (error) {
        console.error('Error loading tickers:', error);
        showError('Failed to load tickers. Please try again later.');
    }
}

async function loadTickerData(ticker) {
    try {
        // Update UI
        document.getElementById('companyName').textContent = ticker;
        document.getElementById('companyDescription').textContent = companyNames[ticker] || 'No description available';
        
        // Show loading state
        showLoading(true);

        const response = await fetch(`/data/${ticker}_data.csv`);
        if (!response.ok) {
            throw new Error(`Failed to fetch data for ${ticker}: ${response.statusText}`);
        }

        const csvData = await response.text();
        
        // Validate CSV data
        await validateCSVData(csvData, `${ticker}_data.csv`);

        const rows = csvData.split('\n').slice(1).filter(row => row.trim());
        const ohlc = [];
        const volumeData = [];

        rows.forEach((row, index) => {
            const columns = row.split(',');
            
            const cleanNumber = (str) => {
                if (!str) return null;
                const cleaned = str.replace(/"/g, '').replace(/,/g, '');
                const num = parseFloat(cleaned);
                return isNaN(num) ? null : num;
            };

            try {
                if (columns[0] && columns[0].trim()) {
                    const [day, month, year] = columns[0].split('/');
                    const date = new Date(year, month - 1, day).getTime();

                    const open = cleanNumber(columns[1]);
                    const high = cleanNumber(columns[2]);
                    const low = cleanNumber(columns[3]);
                    const close = cleanNumber(columns[4]);
                    const volume = cleanNumber(columns[7]);

                    if (date && open !== null && high !== null && low !== null && close !== null && volume !== null) {
                        // Additional validation
                        if (high < low) {
                            console.warn(`Invalid price data at row ${index + 2}: High (${high}) < Low (${low})`);
                            return;
                        }
                        if (open > high || open < low || close > high || close < low) {
                            console.warn(`Invalid OHLC data at row ${index + 2}`);
                            return;
                        }
                        if (volume < 0) {
                            console.warn(`Negative volume at row ${index + 2}`);
                            return;
                        }

                        ohlc.push([date, open, high, low, close]);
                        volumeData.push([date, volume]);
                    }
                }
            } catch (err) {
                console.warn(`Error processing row ${index + 2}:`, err);
            }
        });

        // Ensure we have enough data
        if (ohlc.length < 2) {
            throw new Error(`Insufficient valid data points for ${ticker}`);
        }

        // Sort data by date
        ohlc.sort((a, b) => a[0] - b[0]);
        volumeData.sort((a, b) => a[0] - b[0]);

        createStockChart(ticker, ohlc, volumeData);
    } catch (error) {
        console.error('Error loading data:', error);
        showError(error.message);
    } finally {
        showLoading(false);
    }
}

function createStockChart(ticker, ohlc, volumeSeriesData) {
    if (stockChart) {
        stockChart.destroy();
    }

    // Ensure data is sorted by date
    ohlc.sort((a, b) => a[0] - b[0]);
    volumeSeriesData.sort((a, b) => a[0] - b[0]);

    stockChart = Highcharts.stockChart('priceChart', {
        chart: {
            animation: false // Might help with initial rendering
        },

        yAxis: [{
            labels: { align: 'right', x: -3 },
            title: { text: 'OHLC' },
            height: '60%',
            lineWidth: 2,
            resize: { enabled: true }
        }, {
            labels: { align: 'right', x: -3 },
            title: { text: 'Volume' },
            top: '65%',
            height: '35%',
            offset: 0,
            lineWidth: 2
        }],

        stockTools: {
            gui: {
                enabled: true,
                buttons: [
                    'indicators',
                    'separator',
                    'simpleShapes',
                    'lines',
                    'crookedLines',
                    'measure',
                    'advanced',
                    'toggleAnnotations',
                    'separator',
                    'verticalLabels',
                    'flags',
                    'separator',
                    'zoomChange',
                    'fullScreen',
                    'typeChange',
                    'separator',
                    'currentPriceIndicator'
                ],
                position: {
                    align: 'left',
                    x: 0,
                    y: 0
                }
            }
        },

        rangeSelector: {
            buttons: [{
                type: 'day',
                count: 7,
                text: '1w'
            }, {
                type: 'month',
                count: 1,
                text: '1m'
            }, {
                type: 'month',
                count: 3,
                text: '3m'
            }, {
                type: 'month',
                count: 6,
                text: '6m'
            }, {
                type: 'ytd',
                text: 'YTD'
            }, {
                type: 'year',
                count: 1,
                text: '1y'
            }, {
                type: 'all',
                text: 'All'
            }],
            selected: 1,
            inputEnabled: true
        },

        series: [{
            type: 'candlestick',
            name: ticker,
            id: 'main',
            data: ohlc
        }, {
            type: 'column',
            name: 'Volume',
            id: 'volume',
            data: volumeSeriesData,
            yAxis: 1
        }],

        tooltip: {
            split: true,
            valueDecimals: 2
        },

        plotOptions: {
            candlestick: {
                color: 'red',
                upColor: 'green'
            },
            column: {
                color: '#404040',
                opacity: 0.5
            }
        }
    });

    // Debug data
    console.log('OHLC data sample:', ohlc.slice(0, 5));
    console.log('Volume data sample:', volumeSeriesData.slice(0, 5));
}

// Helper functions for UI feedback
function showLoading(isLoading) {
    const chart = document.getElementById('priceChart');
    if (isLoading) {
        chart.style.opacity = '0.5';
        chart.style.cursor = 'wait';
    } else {
        chart.style.opacity = '1';
        chart.style.cursor = 'default';
    }
}

function showError(message) {
    // You can implement this based on your UI needs
    alert(message);
}

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('tickerSelect').addEventListener('change', (e) => {
        if (e.target.value) {
            loadTickerData(e.target.value);
        }
    });
    loadTickers();
}); 