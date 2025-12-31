# Requirements Document

## Introduction

Remove the non-functional "Profile" menu item from the user dropdown menu in the main layout header to simplify the user interface and avoid confusion.

## Glossary

- **User_Dropdown**: The dropdown menu that appears when clicking on the user avatar in the header
- **Profile_Menu_Item**: The "Profile" option in the user dropdown that currently navigates to `/profile`
- **MainLayout**: The main layout component containing the header and user dropdown

## Requirements

### Requirement 1

**User Story:** As a user, I want a clean and functional user dropdown menu, so that I only see options that actually work.

#### Acceptance Criteria

1. WHEN a user clicks on their avatar in the header THEN the dropdown SHALL only show "Settings" and "Logout" options
2. WHEN a user views the dropdown menu THEN the "Profile" option SHALL not be visible
3. WHEN a user clicks "Settings" THEN the system SHALL navigate to the settings page
4. WHEN a user clicks "Logout" THEN the system SHALL log out the user and redirect to login page
5. THE dropdown menu SHALL maintain proper visual separation between "Settings" and "Logout" with a divider

### Requirement 2

**User Story:** As a developer, I want clean code without unused navigation routes, so that the codebase is maintainable.

#### Acceptance Criteria

1. THE MainLayout component SHALL not contain any reference to profile navigation
2. THE userMenuItems array SHALL only contain settings and logout items
3. THE code SHALL maintain existing functionality for settings and logout
4. THE component SHALL preserve all existing styling and behavior for remaining menu items