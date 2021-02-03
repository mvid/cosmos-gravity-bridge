pragma solidity ^0.6.6;

import "hardhat/console.sol";
import "@openzeppelin/contracts/math/SafeMath.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/SafeERC20.sol";
import "@uniswap/v2-periphery/contracts/interfaces/IUniswapV2Router02.sol";


contract TestUniswapLiquidity is Ownable {
        address router;

        constructor(address _uni_router) public {
                router = _uni_router;
        }

        function redeemLiquidityETH(
        address token,
        uint liquidity,
        uint amountTokenMin,
        uint amountETHMin,
        address to,
        uint deadline
        ) public onlyOwner {
                IUniswapV2Router02(router).removeLiquidityETH(token,liquidity,amountTokenMin,amountETHMin,to,deadline);
        }


        function redeemLiquidity(
        address tokenA,
        address tokenB,
        uint liquidity,
        uint amountAMin,
        uint amountBMin,
        address to,
        uint deadline
        ) public onlyOwner {
                IUniswapV2Router02(router).removeLiquidity(tokenA,tokenB,liquidity,amountAMin,amountBMin,to,deadline);
        }
}