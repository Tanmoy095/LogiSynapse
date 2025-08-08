# LogiSynapse

LogiSynapse is a shipment service application designed to manage logistics operations, including creating, updating, and canceling shipments, generating shipping labels, and comparing carrier rates. Built with Go and PostgreSQL, it integrates with Shippo’s API for real-time tracking and rate comparison, inspired by systems like Amazon and FedEx.

## Features
- **Create Shipments**: Add new shipments with dynamic package dimensions and tracking details.
- **Update Shipments**: Modify shipment details (e.g., destination, ETA) for shipments in `PRE_TRANSIT` status.
- **Cancel Shipments**: Mark shipments as `CANCELLED` using Shippo’s API for label voiding.
- **Rate Comparison**: Fetch and compare shipping rates from carriers like FedEx and UPS via Shippo’s `/rates` endpoint.
- **Label Generation**: Generate printable shipping labels through Shippo’s API.
- **Real-Time Tracking**: Supports tracking updates, with planned webhook integration for instant status changes.

## Installation and Usage
This repository is primarily for viewing and demonstration purposes. To explore the code or run locally (with permission):
1. Clone the repository: `git clone https://github.com/Tanmoy095/LogiSynapse.git`
2. Install dependencies: `go mod tidy`
3. Configure PostgreSQL and Shippo API keys in a `.env` file (not included).
4. Run the service: `go run .`

**Note**: Unauthorized use or modification is prohibited without explicit permission from the owner.

## License
LogiSynapse is licensed under the [Creative Commons Attribution 4.0 International License (CC BY 4.0)](https://creativecommons.org/licenses/by/4.0/). You may share and adapt the material, provided you:
- Give appropriate credit to Tanmoy095.
- Include a link to the license.
- Indicate if changes were made.
See the [LICENSE](LICENSE) file for details.

## Usage Restrictions
This repository is for viewing only unless explicit permission is granted by Tanmoy095. Unauthorized cloning, copying, distribution, or use of the code is prohibited, except as allowed under the CC BY 4.0 license with proper attribution. Contributions (e.g., pull requests, issues) are not accepted without prior approval.

## Contact
For inquiries or permission requests, contact Tanmoy095 via GitHub.
