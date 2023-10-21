const Dotenv = require('dotenv-webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');
module.exports = {
    mode: process.env.NODE_ENV,
    entry: './web/src/app/index.tsx',
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
            {
                test: /\.css$/i,
                use: ["style-loader", "css-loader"],
            },
        ],
    },
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
    },
    // output: {
    //     filename: 'bundle.js',
    //     path: path.resolve(__dirname, './web/assets'),
    // },
    output: {
        path: path.resolve(__dirname, './web/assets'),
        filename: '[name].[contenthash].js',
        // filename: 'bundle.js',
        sourceMapFilename: '[name].[contenthash].map',
        chunkFilename: '[id].[chunkhash].js'
    },
    devtool: 'source-map',
    optimization: {
        splitChunks: {
            chunks: 'all',
            minChunks: 1,
            cacheGroups: {
                defaultVendors: {
                    test: /[\\/]node_modules[\\/]/,
                    priority: -10,
                    reuseExistingChunk: true
                },
                default: {
                    minChunks: 2,
                    priority: -20,
                    reuseExistingChunk: true
                }
            }
        }
    },
    devServer: {
        port: 3000,
        open: true,
        hot: true
    },
    plugins:
        [
            new HtmlWebpackPlugin({
                template: path.resolve(__dirname, './web/src/index_template.html'),
                // hash: true, // Cache busting
                filename: path.resolve(__dirname, './web/views/index.html'),
                publicPath: '/public/assets/'
            }),
            new Dotenv()
        ]
};
