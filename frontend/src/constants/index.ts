/**
 * Successful request status code
 */
export const SUCCESS_CODE = 200

/**
 * Request contentType
 */
export const CONTENT_TYPE = 'application/json'

/**
 * Request timeout in milliseconds
 */
export const REQUEST_TIMEOUT = 60000

/**
 * Whitelist of routes that should not be redirected
 */
export const NO_REDIRECT_WHITE_LIST = ['/login']

/**
 * Whitelist of routes that should not be reset
 */
export const NO_RESET_WHITE_LIST = ['Redirect', 'Login', 'NoFind', 'Root']
