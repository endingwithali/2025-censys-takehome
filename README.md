 

Project Description
In security and network monitoring, understanding change over time is critical. A port that was closed yesterday might be open today; a new vulnerability might appear on a service that was previously considered safe.

Your task is to build a minimal Host Diff Tool:

Ingest snapshots of a host at different points in time.

Compare any two snapshots of the same host to identify what changed.

Provide a simple UI for uploading snapshots, viewing history, and running comparisons.

Generate a structured diff report that highlights meaningful changes (e.g. ports, services, vulnerabilities, or version info).

This exercise tests how you structure a project, handle data, and present results clearly — while staying within a realistic 4-hour window.

Requirements
The project should demonstrate both frontend and backend components.

The system should allow a user to:
- Upload a host snapshot JSON file.
- See a history of snapshots for a given host.
- Select two snapshots and view what has changed between them.
- The application should store snapshots so that they can be retrieved later.
- The application should handle errors gracefully (e.g. bad input, duplicates, or system failures) without crashing.
- Provide a simple web interface to interact with the system. The design doesn’t need to be polished, just functional.
- Use Golang for the backend implementation. The rest is up to you.

Deliverables
- Code Submission: All source code should be uploaded to a publicly accessible GitHub or Bitbucket account.
- README File: A comprehensive README.md file should be included in the repository with the following:
- Instructions on how to run the project.
- Any assumptions made during development.
- Simple testing instructions (manual or automated).
- A brief description of the implemented AI techniques.
- Future Enhancements List: A list detailing what you would do if given more time to work on this project.

# Evaluation Criteria

We will be assessing the following:

Core Functionality: Does the application meet the requirements (upload, store, list, and compare snapshots with a usable UI)?
Correctness & Reliability: Does the system behave as expected, including handling bad inputs or errors gracefully?
Design Decisions: Are the choices around data storage, APIs, and UI sensible and well-reasoned for the problem?
Code Quality: Is the code organized, readable, and maintainable?
Testing: Are there minimal tests in place (especially for critical logic like comparisons)?
Documentation: Does the README clearly explain how to run the project, what assumptions were made, and what could be improved?
User Experience: Is the UI simple and functional, making it easy to complete the required tasks?
Extensibility: Is the project structured in a way that new features could be added without major rework?

Suggested AI Tools
Cursor
CoPilot
ChatGPT
Claude
Any other AI driven development tool you’d like to use.
Please note: Even though we encourage AI assisted tools, we expect that the you understand and can speak to everything in your submitted solution. 

Time Commitment
We expect this project to take approximately 4 hours to complete. We understand that you may have other commitments, so if you anticipate needing more than 5 business days to complete the project, please inform us of your expected submission date.

Resources
Censys Interview data set: host_snapshots

The attached zip file contains several host files. These files contain real Censys host data, but any sensitive information has been replaced with fictional references. You can use these snapshots for testing your solution.

Each file name is formatted as such: `host_<ip>_<timestamp>.json`

IP is the host’s address (e.g., 203.0.113.10).

Timestamp indicates when the snapshot was taken (ISO-8601, with colons replaced by dashes for filenames).


Example: host_125.199.235.74_2025-09-10T03-00-00Z.json

There are three different hosts, each with multiple snapshots taken at different times. Together they illustrate:

- Ports being opened or closed.
- Services being added, removed, or updated.
- Vulnerabilities (CVEs) appearing or being remediated.
- Software versions changing.

You can use these files to test your solution by uploading them, browsing snapshot history, and generating diffs.

Censys Data definitions to learn more about hosts and their structure: https://docs.censys.com/docs/platform-host-dataset#/

Clarification and Support
Please do not hesitate to ask any questions for clarification. We are here to support you throughout this process.

