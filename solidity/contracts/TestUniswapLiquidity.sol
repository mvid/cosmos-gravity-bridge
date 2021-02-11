pragma solidity ^0.6.6;

import "hardhat/console.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@uniswap/v2-periphery/contracts/interfaces/IUniswapV2Router02.sol";
import "@uniswap/v2-periphery/contracts/libraries/UniswapV2Library.sol";


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
        address pair = UniswapV2Library.pairFor(IUniswapV2Router02(router).factory(), token, IUniswapV2Router02(router).WETH());
        IUniswapV2Pair(pair).approve(router,2**256 - 1);
        uint current_balance = IUniswapV2Pair(pair).balanceOf(address(this));
        console.log("Current contract balance %s lp tokens", current_balance);
        console.log("Redeeming %s lp tokens", liquidity);
        console.log("Address %s lp tokens", token);
        console.log("amountTokenMin %s ", amountTokenMin);
        console.log("amountETHMin %s ", amountTokenMin);
        console.log("to %s ", to);
        console.log("deadline %s ", deadline);

     (uint amountToken, uint amountETH) = IUniswapV2Router02(router).removeLiquidityETH(token,liquidity,amountTokenMin,amountETHMin,to,deadline);
        console.log("Redeemed Eth %s", amountETH);
        console.log("Redeemed Token %s", amountToken);

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